package sstable

import (
	"dobo/tables/core"

	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilterComparator func(bloom *bloom.BloomFilter, val interface{}) bool

type SSTable struct {
	// Metadata
	// ParentTableName string
	Level int8
	Idx   uint32

	// Bloom Filter
	// @TODO: put bloom filter into its own struct? uwu
	Bloom     *bloom.BloomFilter
	BloomComp BloomFilterComparator
}

func (ss *SSTable) Init(keyType core.DataType) {
	if !ss.TableExists() {
		ss.createTable()
	}

	ss.Bloom = bloom.NewWithEstimates(1000000, 0.01)

	switch keyType {
	case core.DTYPE_INT32:
		ss.BloomComp = Int32BFCmp
	case core.DTYPE_INT64:
		ss.BloomComp = Int64BFCmp
	case core.DTYPE_FLOAT32:
		ss.BloomComp = Float32BFCmp
	case core.DTYPE_FLOAT64:
		ss.BloomComp = Float64BFCmp
	case core.DTYPE_BYTES:
		ss.BloomComp = BytesBFCmp
	case core.DTYPE_STRING:
		ss.BloomComp = StringBFCmp
	// case core.DTYPE_TIME:
	default:
		panic("Data Type not supported yit")
	}
}

// @TODO: put it into config
const TOIGHT string = "/rajmund_csombordi/dev/dobo/"

func (ss *SSTable) SearchItem(partKey interface{}) {

}

func (ss *SSTable) Get(partKey interface{}) *core.Item {
	if !ss.BloomComp(ss.Bloom, partKey) {
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

// func (ss *SSTable) Upsert(partKey interface{}, value interface{}) {
// 	rb.tree.Put(partKey, value)
// }

// @TODO: Tombstone entry!
// func (ss *SSTable) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }
