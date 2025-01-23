package sstable

import (
	"cmp"
	"dobo/lsm/core"
)

type IndexSummary[P cmp.Ordered] struct {
	MinKey         P
	MinBlockOffset int32
	MaxKey         P
	MaxBlockOffset int32

	Id              int16
	PartKeyTypeInfo *core.TypeInfo
}
