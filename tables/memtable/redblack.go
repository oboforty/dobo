package memtable

import (
	core "dobo/tables/core"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	cmp "github.com/emirpasic/gods/utils"
)

// MemTable implementing Red-Black Balanced Trees
type RBMemT struct {
	tree *rbt.Tree
}

// type Item struct {
// 	interface{}
// }
// GET, QUERY - bulk query, can filter,
// PUT/UPSERT, UPDATE - partial update, can be bulk
// REMOVE - soft delete, DELETE - hard delete,
// CREATE - fails if exists

// CQRS: (GET, QUERY), (CREATE, UPSERT, UPDATE, REMOVE, DELETE), (BALANCE-INDEX, SETCONFIG, CREATE-TABLE, DELETE-TABLE, REPLICATE-TABLE, CREATE-PARTITION, DELETE-PARTITION, REPARTITION-TABLE)

func (rb *RBMemT) Init(keyType core.DataType) {
	// I decided not to use generics as a generic cmp function with interface{} would mean performance reduction
	// func (rb *RBMemT) Get(partKey interface{}, sortKey interface{}) interface{}

	switch keyType {
	case core.DTYPE_INT32:
		rb.tree = &rbt.Tree{Comparator: cmp.Int32Comparator}
	case core.DTYPE_INT64:
		rb.tree = &rbt.Tree{Comparator: cmp.Int64Comparator}
	case core.DTYPE_FLOAT32:
		rb.tree = &rbt.Tree{Comparator: cmp.Float32Comparator}
	case core.DTYPE_FLOAT64:
		rb.tree = &rbt.Tree{Comparator: cmp.Float64Comparator}
	// case core.DTYPE_BYTES:
	case core.DTYPE_STRING:
		rb.tree = &rbt.Tree{Comparator: cmp.StringComparator}
	// case core.DTYPE_TIME:
	// 	rb.tree = &rbt.Tree{Comparator: cmp.TimeComparator}
	default:
		panic("Data Type not supported yit")
	}
}

func (rb *RBMemT) Get(partKey interface{}) *core.Item {
	node := rb.tree.GetNode(partKey)

	if node != nil {
		return &core.Item{
			PartKey: node.Key,
			Value:   node.Value,

			FoundIn:      core.FOUND_AT_MEM,
			FoundSSLevel: -1,
		}
	}

	return nil
}

func (rb *RBMemT) Upsert(partKey interface{}, value interface{}) {
	rb.tree.Put(partKey, value)
}

// @TODO: Tombstone entry!
// func (rb *RBMemT) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }
