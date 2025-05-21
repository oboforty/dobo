package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"iter"
	"path/filepath"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable/ioutils"
)

type IterableTable[P core.PartKeyTypes] interface {
	Len() uint32
	AvgKeySize() uint32
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

func (ss *SSTable[P]) WriteToDisc(table IterableTable[P]) error {
	core.EnsurePath(filepath.Dir(ss.FileBase()))

	// SSTable.New has created a bloomtree, but create it again, now with an estimate for items!
	ss.bloom = bloom.New(bloom.CfgBloomFilter{
		MaxItems:          table.Len(),
		FalsePositiveRate: ss.bloom.FalsePositiveRate(),
	})

	// @TODO: option for .dat file to be uncompressed?
	dat_file, err := ioutils.NewBlockWriter(ss.FileBase()+".dat", ss.compressionBlockSize, true)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file.Close()

	idx_file, err := ioutils.NewBlockWriter(ss.FileBase()+".idx", ss.compressionBlockSize, false)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer idx_file.Close()

	// summary really should be empty always at this step
	ss.sparseIndex = make([]SparseIndex[P], 0)

	// Summary file info
	// index file entries / summary file entries (needed to abide the max size)
	summary_entries := ss.MaxSumSize / table.AvgKeySize()
	summaryInterval := (table.Len() / ss.MinIdxInterval) / summary_entries
	if summaryInterval > 2 {
		ss.Metadata["sparse_type"] = "sum"

		return fmt.Errorf("@TODO: save summary index entries")

		// @TODO: trim summaries by the summary interval
		// ss.summaries
	} else {
		// there's no point in creating index entries for the summary file
		// DB can just load the index file in memory instead.
		ss.Metadata["sparse_type"] = "idx"
		// ss.summaries
	}

	var recordsWritten uint32

	// tree is iterated in partition key order!
	for node := range table.ItemIterator() {
		// Get current start of block
		keyLength := core.GetSizeTypeInfo(node.PartKey, &ss.partKeyTypeInfo)

		if recordsWritten%ss.MinIdxInterval == 0 {

			// Write Index File (3 int32 + the dynamic sized key itself)
			blockOffset, interBlockOffset := dat_file.GetOffsets()

			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, keyLength)
			binary.Write(buf, binary.BigEndian, node.PartKey)
			binary.Write(buf, binary.BigEndian, blockOffset)
			binary.Write(buf, binary.BigEndian, interBlockOffset)
			_, err := idx_file.Write(buf.Bytes())
			if err != nil {
				// @TODO: handle remove SSTables & restore from WAL
				return err
			}

			// we write ALL index entries to summary (which is supposed to be a sparse index over the index file itself!)
			// which is okay to do here, because `table` is already a memory component
			// (we might need MemTable x 2 space available for memory, tho)
			// indexFileOffset, _ := idx_file.GetOffsets()
			idx := SparseIndex[P]{
				Id:           uint32(len(ss.sparseIndex)),
				StartPartKey: node.PartKey,
			}

			if summaryInterval > 2 {
				wof, idxOffset := idx_file.GetOffsets()
				println("@@ TODO @@ ", wof, idxOffset)

				// points to .idx file
				idx.BlockOffset = idxOffset
				// idx.InterBlockOffset =
			} else {
				// points to .dat file
				idx.BlockOffset = blockOffset
				idx.InterBlockOffset = interBlockOffset
			}

			ss.sparseIndex = append(ss.sparseIndex, idx)
		}

		// Write Data file
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, keyLength)
		binary.Write(buf, binary.BigEndian, node.PartKey)
		binary.Write(buf, binary.BigEndian, uint32(len(node.Value)))
		if node.Value != nil {
			buf.Write(node.Value)
		}
		_, err = dat_file.Write(buf.Bytes())
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			return err
		}

		// Write Bloom Filter
		ss.bloom.Add(node.PartKey)

		recordsWritten += 1
	}

	// Write Bloom to disc
	ss.bloom.WriteToDisc(ss.FileBase() + ".bf")

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	WriteSummaryFile(ss.FileBase()+".sum", ss, summaryInterval > 2)

	// post write validations, just to double check things
	if recordsWritten != table.Len() {
		// non-fatal error, but it should be concerning
		return fmt.Errorf("[SST] Write final size mismatch: %d != %d", recordsWritten, table.Len())
	}

	// @TODO: safeguard: measure estimated index & summary size VS actual entries written (in loop) ?

	return nil
}
