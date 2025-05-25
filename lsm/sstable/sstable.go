package sstable

import (
	"path/filepath"
	"strconv"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
)

type BloomFilter interface {
	Add(val interface{}) error
	Test(val interface{}) (bool, error)
	Merge(val interface{}) error

	FalsePositiveRate() float64
	MaxItems() uint32

	WriteToDisc(string) error
	LoadFromDisc(string) error
}

type SSTable[P core.PartKeyTypes] struct {
	GenerationId    int
	partKeyTypeInfo core.TypeInfo
	tablePath       string

	Metadata    map[string]string
	Statistics  map[string]int
	sparseIndex []SparseIndex[P]

	compressionBlockSize uint32
	MinIdxInterval       uint32
	MaxIdxInterval       uint32
	MaxSumSize           uint32

	bloom BloomFilter
}

type CfgSSTable struct {
	DBPath               string `toml:"path" json:"path"`
	CompressionBlockSize uint32 `toml:"block_size" json:"block_size"`
	MinIdxInterval       uint32 `toml:"min_index_interval"`
	MaxIdxInterval       uint32 `toml:"max_index_interval"`
	MaxSumSize           uint32 `toml:"max_summary_file_size"`

	BloomFilter bloom.CfgBloomFilter `toml:"bloom" json:"bloom"`
}

func New[P core.PartKeyTypes](cfg *CfgSSTable, tableName string, pkt core.TypeInfo, id int) *SSTable[P] {
	ss := &SSTable[P]{
		partKeyTypeInfo: pkt,
		tablePath:       filepath.Join(cfg.DBPath, tableName),
		GenerationId:    id,

		compressionBlockSize: cfg.CompressionBlockSize,
		MinIdxInterval:       cfg.MinIdxInterval,
		MaxIdxInterval:       cfg.MaxIdxInterval,
		MaxSumSize:           cfg.MaxSumSize,

		Metadata:   make(map[string]string),
		Statistics: make(map[string]int),

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

func (ss *SSTable[P]) Get(searchKey P) *core.ItemQuery[P] {
	ok, err := ss.bloom.Test(searchKey)

	if err != nil {
		panic(err)
		// @TODO: log? binary errors?
	}

	if !ok {
		return nil
	}

	siHit := BinSearchSparseIndexMemory(searchKey, ss.sparseIndex)

	if siHit == nil {
		return nil
	}

	if ss.Metadata["sparse_type"] == "sum" {
		// the binary search only gave the block offset within the 2nd order (.idx) index file
		// Now we search thas file to get the .dat file's block offsets
		siHit, err = SearchIndexFile(ss.FileBase()+".idx", searchKey, siHit.GetBlockOffset())

		if err != nil {
			// @TODO: log errors?
			panic(err)
			// return nil
		}
	}

	if err != nil {
		// @TODO: log errors?
		panic(err)
	}

	if siHit == nil {
		return nil
	}
	println("@ ", searchKey, " idx file hit: ", siHit.GetBlockOffset(), siHit.GetInterBlockOffset())

	datHit, err := SearchDataFileGzipBlock(
		ss.FileBase()+".dat",
		int32(siHit.GetBlockOffset()),
		int32(siHit.GetInterBlockOffset()),
		searchKey,
	)
	if err != nil {
		// @TODO: log errors?
		panic(err)
	}

	if datHit == nil {
		// @TODO: Panic?
		return nil
	} else if datHit.Value == nil {
		return &core.ItemQuery[P]{
			PartKey:    searchKey,
			Value:      nil,
			FoundIn:    core.FOUND_AT_SS,
			FoundSSIdx: uint32(ss.GenerationId),
			Deleted:    true,
		}
	}

	return &core.ItemQuery[P]{
		PartKey: searchKey,
		Value:   datHit.Value,
		FoundIn: core.FOUND_AT_SS,
	}
}

func (ss *SSTable[P]) LoadFromDisc() error {
	// load bloom filter
	if err := ss.bloom.LoadFromDisc(ss.FileBase() + ".bf"); err != nil {
		return err
	}

	if err := ReadSummaryFile(ss); err != nil {
		return err
	}

	return nil
}
