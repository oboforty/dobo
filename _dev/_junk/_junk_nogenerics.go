
// type Orderable interface {
// 	~int | ~int8 | ~int16 | ~int32 | ~int64 |
// 		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
// 		~float32 | ~float64 |
// 		~string
// }

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

// func (ss *SSTable[P]) Init(keyType core.DataType) {
// 	if !ss.TableExists() {
// 		ss.createTable()
// 	}

