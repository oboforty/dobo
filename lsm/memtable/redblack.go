package memtable

import (
	"cmp"
	"iter"
	"unsafe"

	"github.com/oboforty/dobo/lsm/core"

	rbt "github.com/emirpasic/gods/v2/trees/redblacktree"
	// cmp "github.com/emirpasic/gods/v2/utils"
)

// Calculate the size of a RB Tree Node, without the Key & Value sizes
const RB_NODE_PTRS_SIZE = uint32(unsafe.Sizeof(rbt.Node[int8, int8]{}) - (unsafe.Sizeof(int8(0)) * 2))

// MemTable implementing Red-Black Balanced Trees
type RBMemT[P cmp.Ordered] struct {
	MemT[P]

	tree *rbt.Tree[P, []byte]
}

func NewRedBlack[P cmp.Ordered](cfg *CfgMemtable, partKeyTypeInfo *core.TypeInfo) *RBMemT[P] {
	rb := &RBMemT[P]{
		MemT: MemT[P]{
			partKeyTypeInfo: partKeyTypeInfo,
			maxSize:         cfg.MaxByteSize,
		},
	}
	rb.tree = rbt.New[P, []byte]()

	return rb
}

func (rb *RBMemT[P]) Get(partKey P) *core.ItemQuery[P] {
	node := rb.tree.GetNode(partKey)

	if node != nil {
		return &core.ItemQuery[P]{
			PartKey: node.Key,
			Value:   node.Value,

			FoundIn:      core.FOUND_AT_MEM,
			FoundSSLevel: -1,
		}
	}

	return nil
}

func (rb *RBMemT[P]) Upsert(item *core.ItemWrite[P]) {
	rb.tree.Put(item.PartKey, item.Value)

	// Calculate memory allocation of item
	// Partition key
	if rb.partKeyTypeInfo.IsDynamicSize {
		rb.byteSize += core.GetSize(item.PartKey)
	} else {
		rb.byteSize += rb.partKeyTypeInfo.StaticSize
	}

	// Value & Node structure size
	rb.byteSize += uint32(len(item.Value)) + RB_NODE_PTRS_SIZE
}

// @TODO: Tombstone entry!
// func (rb *RBMemT) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }

func (rb *RBMemT[P]) Clear() {
	rb.tree.Clear()
	rb.byteSize = 0
}

func (rb *RBMemT[P]) ItemIterator() iter.Seq[*core.ItemQuery[P]] {
	return func(yield func(*core.ItemQuery[P]) bool) {

		it := rb.tree.Iterator()

		for i := 0; it.Next(); i++ {
			node := it.Node()

			item := &core.ItemQuery[P]{
				PartKey: node.Key,
				Value:   node.Value,

				FoundIn:      core.FOUND_AT_MEM,
				FoundSSLevel: -1,
			}

			if !yield(item) {
				return
			}
		}
	}
}
