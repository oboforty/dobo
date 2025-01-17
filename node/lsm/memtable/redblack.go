package memtable

import (
	"cmp"
	"dobo/lsm/core"
	"unsafe"

	rbt "github.com/emirpasic/gods/v2/trees/redblacktree"
	// cmp "github.com/emirpasic/gods/v2/utils"
)

// Calculate the size of a RB Tree Node, without the Key & Value sizes
const RB_NODE_PTRS_SIZE = uint(unsafe.Sizeof(rbt.Node[int8, int8]{}) - (unsafe.Sizeof(int8(0)) * 2))

// MemTable implementing Red-Black Balanced Trees
type RBMemT[P cmp.Ordered, V any] struct {
	PartKeyTypeInfo *core.TypeInfo
	SortKeyTypeInfo *core.TypeInfo
	ValueTypeInfo   *core.TypeInfo

	MaxSize uint

	tree *rbt.Tree[P, V]
}

func NewRedBlack[P cmp.Ordered, V any]() (*RBMemT[P, V], error) {
	pkt, err := core.GetTypeInfo[P]()
	if err != nil {
		return nil, err
	}

	// skt, err := core.GetTypeInfo[S]()
	// if err != nil {
	// 	return nil, err
	// }

	vkt, err := core.GetTypeInfo[V]()
	if err != nil {
		return nil, err
	}

	rb := &RBMemT[P, V]{
		PartKeyTypeInfo: pkt,
		ValueTypeInfo:   vkt,
		MaxSize:         128 * 1024 * 1024, // TODO: config
		tree:            rbt.New[P, V](),
	}

	return rb, nil
}

// func (rb *RBMemT) Init() {
// 	// I decided not to use generics as a generic cmp function with interface{} would mean performance reduction
// 	// func (rb *RBMemT) Get(partKey interface{}, sortKey interface{}) interface{}

// 	switch rb.PartKeyType {
// 	case core.DTYPE_INT32:
// 		rb.tree = &rbt.Tree{Comparator: cmp.Int32Comparator}
// 	case core.DTYPE_INT64:
// 		rb.tree = &rbt.Tree{Comparator: cmp.Int64Comparator}
// 	case core.DTYPE_FLOAT32:
// 		rb.tree = &rbt.Tree{Comparator: cmp.Float32Comparator}
// 	case core.DTYPE_FLOAT64:
// 		rb.tree = &rbt.Tree{Comparator: cmp.Float64Comparator}
// 	// case core.DTYPE_BYTES:
// 	case core.DTYPE_STRING:
// 	// case core.DTYPE_TIME:
// 	// 	rb.tree = &rbt.Tree{Comparator: cmp.TimeComparator}
// 	default:
// 		panic("Data Type not supported yit")
// 	}
// }

func (rb *RBMemT[P, V]) Get(partKey P) *core.ItemQuery[P, V] {
	node := rb.tree.GetNode(partKey)

	if node != nil {
		return &core.ItemQuery[P, V]{
			PartKey: node.Key,
			Value:   node.Value,

			FoundIn:      core.FOUND_AT_MEM,
			FoundSSLevel: -1,
		}
	}

	return nil
}

func (rb *RBMemT[P, V]) Upsert(item *core.ItemWrite[P, V]) {

	if rb.PartKeyTypeInfo.IsDynamicSize {
		rb.PartKeyTypeInfo.InsertedSize += core.GetSize(item.PartKey)
	}

	if rb.ValueTypeInfo.IsDynamicSize {
		rb.ValueTypeInfo.InsertedSize += core.GetSize(item.Value)
	}

	rb.tree.Put(item.PartKey, item.Value)
}

// @TODO: Tombstone entry!
// func (rb *RBMemT) Delete(partKey interface{}) {
// 	rb.tree.Remove(partKey)
// }

func (rb *RBMemT[P, V]) ByteSize() uint {
	var totalSize uint = 0
	length := uint(rb.tree.Size())

	if rb.PartKeyTypeInfo.IsDynamicSize {
		totalSize += rb.PartKeyTypeInfo.InsertedSize
	} else {
		totalSize += rb.PartKeyTypeInfo.StaticSize * length
	}

	if rb.ValueTypeInfo.IsDynamicSize {
		totalSize += rb.ValueTypeInfo.InsertedSize
	} else {
		totalSize += rb.ValueTypeInfo.StaticSize * length
	}

	totalSize += RB_NODE_PTRS_SIZE * length

	return totalSize
}

func (rb *RBMemT[P, V]) IsFull() bool {
	return rb.MaxSize <= rb.ByteSize()
}

// Creates a shallow copy of the struct and clears
func (rb *RBMemT[P, V]) CloneAndClear() *RBMemT[P, V] {
	cloneRB := *rb
	rb.tree.Clear()

	return &cloneRB
}
