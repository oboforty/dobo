package sstable

import (
	"cmp"
	"dobo/lsm/core"
	"encoding/binary"
	"fmt"
	"iter"
	"os"
	"path/filepath"

	"github.com/bits-and-blooms/bloom/v3"
)

type MemTable[P cmp.Ordered] interface {
	ByteSize() uint
	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

type BloomFilterComparator func(bloom *bloom.BloomFilter, val interface{}) bool

type SSTable[P cmp.Ordered] struct {
	// Metadata
	TableName       string
	PartKeyTypeInfo *core.TypeInfo
	BasePath        string
	GenerationId    int
	Level           int8

	// bloom Filter
	// @TODO: put bloom filter into its own struct? + even add interface?
	bloom     *bloom.BloomFilter
	bloomComp BloomFilterComparator
}

func (ss *SSTable[P]) Init() {
	// ss := &SSTable[P]{}

	// @TODO: write our own bloom filters with generics
	ss.bloom = bloom.NewWithEstimates(1000000, 0.01)
}

func (ss *SSTable[P]) GetGenerationId() int {
	return ss.GenerationId
}

func (ss *SSTable[P]) Get(partKey P) *core.ItemQuery[P] {
	if !ss.bloomComp(ss.bloom, partKey) {
		// key is defo not in this table
		return nil
	}

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
	if _, err := os.Stat(ss.BasePath); err != nil {
		err = os.MkdirAll(ss.BasePath, os.ModePerm)

		if err != nil {
			return os.ErrNotExist
		}
	}

	ssfilename := filepath.Join(ss.BasePath, fmt.Sprintf("%s-%d", ss.TableName, ss.GenerationId))

	dat_file_inner, err := os.OpenFile(ssfilename+".db", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dat_file_inner.Close()
	idx_file, err := os.OpenFile(ssfilename+".idx", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer idx_file.Close()

	dat_file := NewCountingWriter(dat_file_inner)

	// tree is iterated in partition key order!
	for node := range memt.ItemIterator() {
		// Write Data file
		// @TODO: strategy between binary, gop, jsonl
		err = binary.Write(dat_file, binary.BigEndian, node.Value)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}
		currentOffset := int32(dat_file.BytesWritten())

		// Write Index File
		err := binary.Write(idx_file, binary.BigEndian, node.PartKey)
		if err != nil {
			// @TODO: handle remove SSTables & restore from WAL
			panic(err)
		}
		err = binary.Write(idx_file, binary.BigEndian, currentOffset)
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
