package memtable

import (
	"iter"
	"unsafe"

	"github.com/oboforty/dobo/lsm/core"
	rbt "github.com/oboforty/dobo/lsm/trees/redblacktree"
)

// Calculate the size of a RB Tree Node, without the Key & Value sizes
const RB_NODE_PTRS_SIZE = uint64(unsafe.Sizeof(rbt.Node[int8, int8]{}) - (unsafe.Sizeof(int8(0)) * 2))

// MemTable implementing Red-Black Balanced Trees
type RBMemT[P core.PartKeyTypes] struct {
	MemT

	tree *rbt.Tree[P, []byte]
}

func NewRedBlack[P core.PartKeyTypes](cfg *CfgMemtable, partKeyTypeInfo *core.TypeInfo) *RBMemT[P] {
	rb := &RBMemT[P]{
		MemT: MemT{
			partKeyTypeInfo:     partKeyTypeInfo,
			maxSize:             cfg.MaxByteSize,
			recordAuxiliarySize: RB_NODE_PTRS_SIZE,
		},
	}
	rb.tree = rbt.NewWith[P, []byte](core.UberComparator)

	return rb
}

func (rb *RBMemT[P]) Get(partKey P) *core.ItemQuery[P] {
	leaf := rb.tree.GetNode(partKey)

	if leaf != nil {

		if leaf.Value == nil {
			// nil means Tombstone (item deleted)
			return &core.ItemQuery[P]{
				PartKey:      leaf.Key,
				FoundIn:      core.FOUND_AT_MEM,
				FoundSSLevel: -1,
				Deleted:      true,
			}
		}

		return &core.ItemQuery[P]{
			PartKey:      leaf.Key,
			Value:        leaf.Value,
			FoundIn:      core.FOUND_AT_MEM,
			FoundSSLevel: -1,
		}
	}

	return nil
}

func (rb *RBMemT[P]) Upsert(item *core.Item[P]) {
	rb.tree.Put(item.PartKey, item.Value)

	// Calculate memory allocation of item
	// Partition key
	if rb.partKeyTypeInfo.IsDynamicSize {
		rb.totalKeySize += uint64(core.GetSize(item.PartKey))
	} else {
		rb.totalKeySize += uint64(rb.partKeyTypeInfo.StaticSize)
	}

	// Value & Node structure size
	rb.totalValueSize += uint64(len(item.Value)) + rb.recordAuxiliarySize
}

// Adds a tombstone entry to RB Tree. Returns true if item has been deleted in memory
func (rb *RBMemT[P]) Delete(partKey P) bool {
	leaf := rb.tree.GetNode(partKey)

	if leaf != nil {
		// set the entry to be a tombstone
		leaf.Value = nil

		return true
	} else {
		// add a tombstone entry
		rb.tree.Put(partKey, nil)

		return false
	}
}

func (rb *RBMemT[P]) ItemIterator() iter.Seq[*core.Item[P]] {
	return func(yield func(*core.Item[P]) bool) {

		it := rb.tree.Iterator()

		for i := 0; it.Next(); i++ {
			node := it.Node()

			item := &core.Item[P]{
				PartKey: node.Key,
				Value:   node.Value,
			}

			if !yield(item) {
				return
			}
		}
	}
}

func (rb *RBMemT[P]) Len() uint32 {
	return uint32(rb.tree.Size())
}

func (rb *RBMemT[P]) Clear() {
	rb.tree.Clear()
	rb.totalValueSize = 0
	rb.totalKeySize = 0
}
