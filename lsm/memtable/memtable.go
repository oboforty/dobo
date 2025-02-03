package memtable

import (
	"cmp"

	"github.com/oboforty/dobo/lsm/core"
)

type CfgMemtableType string

const (
	MEMTYPE_REDBLACK CfgMemtableType = "redblack"
	MEMTYPE_AVL      CfgMemtableType = "avl"
	MEMTYPE_SKIPLIST CfgMemtableType = "skiplist"
)

type CfgMemtable struct {
	Type        CfgMemtableType `toml:"type"`
	MaxByteSize uint32          `toml:"max_size"`
}

type MemT[P cmp.Ordered] struct {
	partKeyTypeInfo *core.TypeInfo

	maxSize  uint32
	byteSize uint32
}

func (m *MemT[P]) ByteSize() uint32 {
	return m.byteSize
}

func (m *MemT[P]) IsFull() bool {
	return m.maxSize <= m.byteSize
}
