package sstable

import (
	"cmp"
	"dobo/lsm/core"
	"dobo/lsm/utils"
	"encoding/binary"
	"fmt"
	"iter"
	"path/filepath"
	"strconv"
)

type IterableTable[P cmp.Ordered] interface {
	// ByteSize() uint
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

type SSTable[P cmp.Ordered] struct {
	GenerationId int

	partKeyTypeInfo      core.TypeInfo
	dbpath               string
	dataSerialization    core.DynamicValueSerialization
	compressionBlockSize int

	bloom     *bloomFilter
	summaries []IndexSummary[P]
}

type CfgSSTable struct {
	BasePath             string                         `toml:"base_path"`
	DataSerialization    core.DynamicValueSerialization `toml:"data_serialization"`
	CompressionBlockSize int                            `toml:"compression_block_size"`

	BloomFilter CfgBloomFilter `toml:"bloom_filter"`
}

func New[P cmp.Ordered](cfg *CfgSSTable, tableName string, pkt core.TypeInfo, id int) *SSTable[P] {
	dbpath := filepath.Join(cfg.BasePath, tableName, strconv.Itoa(id))
	utils.EnsurePath(filepath.Dir(dbpath))

	ss := &SSTable[P]{
		partKeyTypeInfo:      pkt,
		dbpath:               dbpath,
		GenerationId:         id,
		dataSerialization:    cfg.DataSerialization,
		compressionBlockSize: cfg.CompressionBlockSize,
		// bloom:             newBloomFilter(cfg.BloomFilter),
	}

	return ss
}

func (ss *SSTable[P]) GetGenerationId() int {
	return ss.GenerationId
}

func (ss *SSTable[P]) Get(partKey P) *core.ItemQuery[P] {
	// if !ss.bloomComp(ss.bloom, partKey) {
	// 	// key is defo not in this table
	// 	return nil
	// }
	// 		println("###", keyLength, "OFF:", blockOffset, interBlockOffset, fmt.Sprintf("value:\t %v", oof))

	// @TODO: load summary from dsic if nil!
	var idxRange *IndexSummary[P]
	for _, sum := range ss.summaries {
		if sum.MinKey <= partKey && partKey <= sum.MaxKey {
			idxRange = &sum
		}
	}

	if idxRange == nil {
		return nil
	}
	println("!!! Looking for:", partKey, " Summary Range:", idxRange.Id, "/", len(ss.summaries), idxRange.MinBlockOffset, idxRange.MaxBlockOffset)

	// @TODO: binary search idx file
	oof, err := SearchKeyInFile2(
		ss.dbpath+".idx",
		idxRange.MinBlockOffset,
		idxRange.MaxBlockOffset,
		partKey,
	)
	if err != nil {
		panic(err)
	}
	println("@@@@@", oof)

	// check idx

	// readBlock()

	// try disk IO
	// return &core.Item{
	// 	PartKey: node.Key,
	// 	Value:   node.Value,

	// 	FoundIn:      core.FOUND_AT_SS,
	// 	FoundSSLevel: ss.Level,
	// }

	return nil
}

// func (ss *SSTable[P]) Upsert(partKey interface{}, value interface{}) {
// 	rb.tree.Put(partKey, value)
// }

// @TODO: Tombstone entry!
// func (ss *SSTable[P]) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }

func (ss *SSTable[P]) WriteToDisc(table IterableTable[P]) error {

	// @TODO: option for .dat file to be uncompressed?
	dat_file, err := NewBlockWriter(ss.dbpath+".dat", ss.compressionBlockSize, true)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file.Close()

	// @TODO: separate cfg for block size
	idxBlockSize := ss.compressionBlockSize
	idx_file, err := NewBlockWriter(ss.dbpath+".idx", idxBlockSize, false)
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
				MinBlockOffset:  blockOffset,
			}
		}

		// Write Index File (3 int32 + the dynamic sized key itself)
		idxContent := make([]byte, 0, 4*3+keyLength)
		idxContent = binary.BigEndian.AppendUint32(idxContent, uint32(keyLength))
		idxContent, _ = binary.Append(idxContent, binary.BigEndian, currentKey)
		idxContent = binary.BigEndian.AppendUint32(idxContent, uint32(blockOffset))
		idxContent = binary.BigEndian.AppendUint32(idxContent, uint32(interBlockOffset))

		// oof := make([]byte, 0)
		// oof, _ = binary.Append(oof, binary.BigEndian, currentKey)
		// println("###", keyLength, "OFF:", idxBlockOffset, asd, keyLength)

		_, err := idx_file.Write(idxContent)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}

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
		_, err = dat_file.Write(node.Value)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}

		// asd += 1
		// if asd > 10 {
		// 	break
		// }
	}

	// @TODO: ITT: TEST WITH full range

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
