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
	MaxByteSize uint32          `toml:"max_size" json:"max_size"`
}

type MemT struct {
	partKeyTypeInfo *core.TypeInfo

	maxSize  uint32
	byteSize uint32
}

func (m *MemT) ByteSize() uint32 {
	return m.byteSize
}

func (m *MemT) IsFull() bool {
	return m.maxSize <= m.byteSize
}
