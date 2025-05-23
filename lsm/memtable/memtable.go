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

	maxSize             uint64
	totalValueSize      uint64
	totalKeySize        uint64
	recordAuxiliarySize uint64
}

func (m *MemT) Len() uint32 {
	// must be substituted by MemT implementation
	return 0
}

func (m *MemT) ByteSize() uint64 {
	// implementation calculates the extra bytes needed for one record (e.g. tree pointers)
	return m.TotalKeySize() + m.TotalValueSize() + uint64(m.Len())*uint64(m.recordAuxiliarySize)
}

func (m *MemT) TotalValueSize() uint64 {
	return m.totalValueSize
}

func (m *MemT) TotalKeySize() uint64 {
	return m.totalKeySize
}

func (m *MemT) IsFull() bool {
	return m.ByteSize() >= m.maxSize
}

// func (m *MemT) AvgKeySize() uint32 {
// 	if m.partKeyTypeInfo.IsDynamicSize {
// 		return uint32(m.totalKeySize / uint64(m.Len()))
// 	} else {
// 		return m.partKeyTypeInfo.StaticSize
// 	}
// }
