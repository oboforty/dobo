package memtable

import (
	"github.com/oboforty/dobo/lsm/core"
)

type CfgMemtableType string

const (
	MEMTYPE_REDBLACK CfgMemtableType = "redblack"
	MEMTYPE_AVL      CfgMemtableType = "avl"
	MEMTYPE_SKIPLIST CfgMemtableType = "skiplist"
)

type CfgMemtable struct {
	Type        CfgMemtableType `toml:"type" json:"type"`
	MaxByteSize uint64          `toml:"max_size" json:"max_size"`
}

type MemT struct {
	partKeyTypeInfo *core.TypeInfo

	maxSize        uint64
	totalValueSize uint64
	totalKeySize   uint64
}

func (m *MemT) ByteSize() uint64 {
	return m.totalKeySize + m.totalValueSize
}

func (m *MemT) IsFull() bool {
	return m.ByteSize() >= m.maxSize
}
