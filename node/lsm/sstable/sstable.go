package sstable

import (
	"cmp"
	"dobo/lsm/core"
	"encoding/binary"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strconv"
)

type MemTable[P cmp.Ordered] interface {
	ByteSize() uint
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

type SSTable[P cmp.Ordered] struct {
	GenerationId int

	partKeyTypeInfo      *core.TypeInfo
	dbpath               string
	dataSerialization    core.DynamicValueSerialization
	compressionBlockSize uint32

	bloom *bloomFilter
}

type CfgSSTable struct {
	BasePath             string                         `toml:"base_path"`
	DataSerialization    core.DynamicValueSerialization `toml:"data_serialization"`
	CompressionBlockSize uint32                         `toml:"compression_block_size"`

	BloomFilter CfgBloomFilter `toml:"bloom_filter"`
}

func New[P cmp.Ordered](cfg *CfgSSTable, partKeyTypeInfo *core.TypeInfo, tableName string, id int) *SSTable[P] {

	ss := &SSTable[P]{
		partKeyTypeInfo:   partKeyTypeInfo,
		dbpath:            filepath.Join(cfg.BasePath, tableName, strconv.Itoa(id)),
		GenerationId:      id,
		bloom:             newBloomFilter(cfg.BloomFilter),
		dataSerialization: cfg.DataSerialization,
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

func (ss *SSTable[P]) WriteMemToDisc(memt MemTable[P]) error {

	idx_file, err := os.OpenFile(ss.dbpath+".idx", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer idx_file.Close()

	dat_file, err := NewCompressedWriter(ss.dbpath+".dat", ss.compressionBlockSize)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file.Close()

	// tree is iterated in partition key order!
	for node := range memt.ItemIterator() {

		// Write Data file
		blockOffset, err := dat_file.Write(node.Value)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}

		// @TODO: handle string and []byte keys! binary uses Reflect!
		// Write Index File
		err = binary.Write(idx_file, binary.BigEndian, node.PartKey)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}
		err = binary.Write(idx_file, binary.BigEndian, blockOffset)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}
	}

	// todo: write summary file
	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	return nil
}
