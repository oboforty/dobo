package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable/ioutils"
)

func (ss *SSTable[P]) WriteToDisc(table IterableTable[P]) error {

	// @TODO: option for .dat file to be uncompressed?
	dat_file, err := ioutils.NewBlockWriter(ss.dbpath+".dat", ss.compressionBlockSize, true)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file.Close()

	// @TODO: separate cfg for block size
	idxBlockSize := ss.compressionBlockSize
	idx_file, err := ioutils.NewBlockWriter(ss.dbpath+".idx", idxBlockSize, false)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer idx_file.Close()

	// idx block offsets for summary file
	var idxBlockOffsetPrevious int32 = 0

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
		binary.Write(buf, binary.BigEndian, uint32(keyLength))
		binary.Write(buf, binary.BigEndian, currentKey)
		binary.Write(buf, binary.BigEndian, uint32(blockOffset))
		binary.Write(buf, binary.BigEndian, uint32(interBlockOffset))
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

		// asd += 1
		// if asd > 10 {
		// 	break
		// }
	}

	// @TODO: NEM JO A LAST OFFSET
	// write last summary entry
	currentSummary.MaxKey = currentKey
	currentSummary.MaxBlockOffset = idxBlockOffsetPrevious

	if currentSummary.MaxBlockOffset == 0 {
		println("!!! ERROR: APPEND file length as last summary item's size!")
		currentSummary.MaxBlockOffset = 99999
	}
	ss.summaries = append(ss.summaries, *currentSummary)

	// @TOOD: WRITE TO DISC

	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	return nil
}
