package sstable

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"iter"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable/ioutils"
)

type IterableTable[P cmp.Ordered] interface {
	Size() uint32
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

func (ss *SSTable[P]) WriteToDisc(table IterableTable[P]) error {
	// SSTable.New has created a bloomtree, but create it again, now with an estimate for items!
	ss.bloom = bloom.New(bloom.CfgBloomFilter{
		MaxItems:          table.Size(),
		FalsePositiveRate: ss.bloom.FalsePositiveRate(),
	})

	// @TODO: option for .dat file to be uncompressed?
	dat_file, err := ioutils.NewBlockWriter(ss.tablePath+".dat", ss.compressionBlockSize, true)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file.Close()

	idxBlockSize := ss.compressionBlockSize
	idx_file, err := ioutils.NewBlockWriter(ss.tablePath+".idx", idxBlockSize, false)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer idx_file.Close()

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
			panic(err)
		}

		// Write Summary file
		if idxBlockOffsetPrevious != idxBlockOffset {
			currentSummary.MaxKey = currentKey
			currentSummary.MaxBlockOffset = idxBlockOffset

			// Write Summary file for each index block (only)
			// println("@@", currentSummary.MinKey, currentSummary.MaxKey, currentSummary.BlockOffset, "   ", idxBlockOffset, idxInterBlockOffset)

			// @TOOD: WRITE TO DISC

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

	// @TODO: NEM JO A LAST OFFSET
	// 				write last summary entry
	currentSummary.MaxKey = currentKey
	currentSummary.MaxBlockOffset = idxBlockOffsetPrevious

	if currentSummary.MaxBlockOffset == 0 {
		println("!!! ERROR: APPEND file length as last summary item's size!")
		currentSummary.MaxBlockOffset = 99999
	}
	ss.summaries = append(ss.summaries, *currentSummary)

	// Write Bloom to disc
	ss.bloom.WriteToDisc(ss.tablePath + ".bf")

	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	if totalItems != table.Size() {
		// non-fatal error, but it should be concerning
		return fmt.Errorf("[SST] Write final size mismatch: %d != %d", totalItems, table.Size())
	}

	return nil
}
