package sstable

import (
	"cmp"
	"iter"
	"path/filepath"
	"strconv"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/utils"
)

type IterableTable[P cmp.Ordered] interface {
	// ByteSize() uint
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

type SSTable[P cmp.Ordered] struct {
	GenerationId int
	Statistics   map[string]float32

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

type IndexSummary[P cmp.Ordered] struct {
	MinKey         P
	MinBlockOffset int32
	MaxKey         P
	MaxBlockOffset int32

	Id              int16
	PartKeyTypeInfo *core.TypeInfo
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

	khit, err := SearchOffsetInIndexFile(
		ss.dbpath+".idx",
		idxRange.MinBlockOffset,
		idxRange.MaxBlockOffset,
		partKey,
	)
	if err != nil {
		panic(err)
	}
	if khit == nil {
		// @TODO: Panic?
		return nil
	}

	vhit, err := SearchDataFileGzipBlock(
		ss.dbpath+".dat",
		int32(khit.BlockOffset),
		int32(khit.InterBlockOffset),
		partKey,
	)
	if err != nil {
		panic(err)
	}
	if vhit == nil {
		// @TODO: Panic?
		return nil
	}

	return &core.ItemQuery[P]{
		PartKey: partKey,
		Value:   vhit.Value,

		FoundIn: core.FOUND_AT_SS,
		// FoundSSLevel: ss.Level,
	}

	return nil
}

// func (ss *SSTable[P]) Upsert(partKey interface{}, value interface{}) {
// 	rb.tree.Put(partKey, value)
// }

// @TODO: Tombstone entry!
// func (ss *SSTable[P]) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }
