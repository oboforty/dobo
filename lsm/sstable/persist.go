package sstable

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"iter"
	"os"
	"path/filepath"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable/ioutils"
)

type IterableTable[P cmp.Ordered] interface {
	Size() uint32
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

func (ss *SSTable[P]) WriteToDisc(table IterableTable[P]) error {
	core.EnsurePath(filepath.Dir(ss.FileBase()))

	// SSTable.New has created a bloomtree, but create it again, now with an estimate for items!
	ss.bloom = bloom.New(bloom.CfgBloomFilter{
		MaxItems:          table.Size(),
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

	sum_file, err := os.OpenFile(ss.FileBase()+".sum", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer sum_file.Close()

	// @TODO: add Stats in sum file
	sum_file.WriteString("------")

	// idx block offsets for summary file
	var idxBlockOffsetPrevious uint32 = 0
	var totalItems uint32 = 0

	// estimate & reserve summaries
	// ss.summaries = make([]IndexSummary[P], 0, )
	var currentSummary *IndexSummary[P]
	var currentKey P

	// tree is iterated in partition key order!
	for node := range table.ItemIterator() {
		// Get current start of block
		blockOffset, interBlockOffset := dat_file.GetOffsets()
		idxBlockOffset, _ := idx_file.GetOffsets()
		keyLength := core.GetSizeTypeInfo(node.PartKey, &ss.partKeyTypeInfo)
		currentKey = node.PartKey

		if currentSummary == nil {
			// reserve new IdxSummary
			currentSummary = &IndexSummary[P]{
				Id:              int16(len(ss.summaries)),
				PartKeyTypeInfo: &ss.partKeyTypeInfo,
				MinKey:          currentKey,
				MinBlockOffset:  idxBlockOffset,
			}
		}

		// Write Index File (3 int32 + the dynamic sized key itself)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, keyLength)
		binary.Write(buf, binary.BigEndian, currentKey)
		binary.Write(buf, binary.BigEndian, blockOffset)
		binary.Write(buf, binary.BigEndian, interBlockOffset)
		_, err := idx_file.Write(buf.Bytes())
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			return err
		}

		// Write Summary file for each index block (only)
		if idxBlockOffsetPrevious != idxBlockOffset {
			currentSummary.MaxKey = currentKey
			currentSummary.MaxBlockOffset = idxBlockOffset

			// @TOOD: WRITE TO DISC
			buf = new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, currentSummary.MinKey)
			binary.Write(buf, binary.BigEndian, currentSummary.MaxKey)
			binary.Write(buf, binary.BigEndian, currentSummary.MinBlockOffset)
			binary.Write(buf, binary.BigEndian, currentSummary.MaxBlockOffset)
			_, err = sum_file.Write(buf.Bytes())
			if err != nil {
				return err
			}

			ss.summaries = append(ss.summaries, *currentSummary)
			currentSummary = nil
			idxBlockOffsetPrevious = idxBlockOffset
		}

		// Write Data file
		buf = new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(node.Value)))
		buf.Write(node.Value)
		_, err = dat_file.Write(buf.Bytes())
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}

		// Write Bloom Filter
		ss.bloom.Add(currentKey)

		totalItems += 1
	}

	if currentSummary.MaxBlockOffset != 0 {
		println("@@@@ THIS SHOULD BE 0: ", currentSummary.MaxBlockOffset)
	}

	// Write last summary entry -- calc upper bound of last
	// @TODO: figure out why these two equal to file's size xD
	idxBlockOffset, idxInterBlockOffset := idx_file.GetOffsets()
	currentSummary.MaxBlockOffset = idxBlockOffset + idxInterBlockOffset
	currentSummary.MaxKey = currentKey
	ss.summaries = append(ss.summaries, *currentSummary)

	// Write Bloom to disc
	ss.bloom.WriteToDisc(ss.FileBase() + ".bf")

	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	if totalItems != table.Size() {
		// non-fatal error, but it should be concerning
		return fmt.Errorf("[SST] Write final size mismatch: %d != %d", totalItems, table.Size())
	}

	return nil
}
