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

	sum_file, err := NewSummaryWriter[P](ss.FileBase() + ".sum")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer sum_file.Close()

	sum_file.WriteMetadata(nil, nil)

	// @TODO: add Stats in sum file

	// idx block offsets for summary file
	var idxBlockOffsetPrevious uint32 = 0
	var totalItems uint32 = 0
	var currentKey P

	// tree is iterated in partition key order!
	for node := range table.ItemIterator() {
		// Get current start of block
		blockOffset, interBlockOffset := dat_file.GetOffsets()
		idxBlockOffset, _ := idx_file.GetOffsets()
		keyLength := core.GetSizeTypeInfo(node.PartKey, &ss.partKeyTypeInfo)
		currentKey = node.PartKey

		if sum_file.Empty() {
			sum_file.StartRegion(len(ss.summaries), keyLength, currentKey, idxBlockOffset)
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
			sum, err := sum_file.StopRegion(currentKey, idxBlockOffset)
			if err != nil {
				return err
			}

			ss.summaries = append(ss.summaries, *sum)
			idxBlockOffsetPrevious = idxBlockOffset
		}

		// Write Data file
		buf = new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(node.Value)))
		if node.Value != nil {
			buf.Write(node.Value)
		}
		_, err = dat_file.Write(buf.Bytes())
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}

		// Write Bloom Filter
		ss.bloom.Add(currentKey)

		totalItems += 1
	}

	err = sum_file.AssertBlocksetOK()
	if err != nil {
		return err
	}

	// Write last summary entry -- calc upper bound of last
	// @TODO: figure out why these two equal to file's size xD
	idxBlockOffset, idxInterBlockOffset := idx_file.GetOffsets()
	sum, err := sum_file.StopRegion(currentKey, idxBlockOffset+idxInterBlockOffset)
	if err != nil {
		return err
	}
	ss.summaries = append(ss.summaries, *sum)

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
