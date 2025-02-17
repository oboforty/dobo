package sstable

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
)

type BloomFilter interface {
	Add(val interface{}) error
	Test(val interface{}) (bool, error)
	FalsePositiveRate() float64
	WriteToDisc(string) error
	LoadFromDisc(string) error
}

type SSTable[P core.PartKeyTypes] struct {
	GenerationId    int
	partKeyTypeInfo core.TypeInfo
	Statistics      map[string]float32
	summaries       []IndexSummary[P]

	tablePath            string
	compressionBlockSize uint32

	bloom BloomFilter
}

type CfgSSTable struct {
	DBPath               string `toml:"path" json:"path"`
	CompressionBlockSize uint32 `toml:"block_size" json:"block_size"`

	BloomFilter bloom.CfgBloomFilter `toml:"bloom" json:"bloom"`
}

type IndexSummary[P core.PartKeyTypes] struct {
	Id             int16
	MinKey         P
	MinBlockOffset uint32
	MaxKey         P
	MaxBlockOffset uint32

	// PartKeyTypeInfo *core.TypeInfo
}

func New[P core.PartKeyTypes](cfg *CfgSSTable, tableName string, pkt core.TypeInfo, id int) *SSTable[P] {
	ss := &SSTable[P]{
		partKeyTypeInfo:      pkt,
		tablePath:            filepath.Join(cfg.DBPath, tableName),
		GenerationId:         id,
		compressionBlockSize: cfg.CompressionBlockSize,

		// Bloom filter doesn't really get created here,
		// as it's actually created after first disc write
		bloom: bloom.New(cfg.BloomFilter),
	}

	return ss
}

func (ss *SSTable[P]) GetGenerationId() int {
	return ss.GenerationId
}

func (ss *SSTable[P]) FileBase() string {
	return filepath.Join(ss.tablePath, "g"+strconv.Itoa(ss.GenerationId))
}

func (ss *SSTable[P]) Get(partKey P) *core.ItemQuery[P] {
	ok, err := ss.bloom.Test(partKey)

	if err != nil {
		// @TODO: log? binary errors?
	}

	if !ok {
		return nil
	}

	// @TODO: load summary from dsic if nil!

	var idxRange *IndexSummary[P]
	for _, sum := range ss.summaries {
		if core.UberComparator(sum.MinKey, partKey) == -1 && core.UberComparator(partKey, sum.MaxKey) == -1 {
			idxRange = &sum
		}
	}

	if idxRange == nil {
		return nil
	}

	khit, err := SearchOffsetInIndexFile(
		ss.FileBase()+".idx",
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
		ss.FileBase()+".dat",
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
	} else if vhit.Value == nil {
		return &core.ItemQuery[P]{
			PartKey:    partKey,
			Value:      nil,
			FoundIn:    core.FOUND_AT_SS,
			FoundSSIdx: uint32(ss.GenerationId),
			Deleted:    true,
		}
	}

	return &core.ItemQuery[P]{
		PartKey: partKey,
		Value:   vhit.Value,
		FoundIn: core.FOUND_AT_SS,
	}
}

// func (ss *SSTable[P]) Upsert(partKey interface{}, value interface{}) {
// 	rb.tree.Put(partKey, value)
// }

// @TODO: Tombstone entry!
// func (ss *SSTable[P]) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }

func (ss *SSTable[P]) LoadFromDisc() error {
	file, err := os.Open(ss.FileBase() + ".sum")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// load bloom filter
	ss.bloom.LoadFromDisc(ss.FileBase() + ".bf")

	return ReadSummaryFile(file, ss)
}
