package core

// type Orderable interface {
// 	~int | ~int8 | ~int16 | ~int32 | ~int64 |
// 		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
// 		~float32 | ~float64 |
// 		~string
// }

type DataType uint8

const (
	DTYPE_INT32 DataType = iota
	DTYPE_INT64
	DTYPE_FLOAT32
	DTYPE_FLOAT64
	DTYPE_BYTES // NOT SUPPORTED YET
	DTYPE_STRING
	DTYPE_TIME // NOT SUPPORTED YET
)

type TableMetadata struct {
	Name string

	TokenMin int
	TokenMax int

	PartKeyType DataType
	SortKeyType DataType
}

type FindStatus = uint8

const (
	NOT_FOUND FindStatus = iota
	FOUND_AT_MEM
	FOUND_AT_BLOOM
	FOUND_AT_SS
)

type Item struct {
	PartKey any // Orderable
	SortKey any // Orderable
	Value   any

	// todo: put these into metadata? or we'll put them at api json lvl?
	FoundIn      FindStatus
	FoundSSLevel int8
	FoundSSIdx   uint32
}
