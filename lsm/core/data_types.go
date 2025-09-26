package core

import (
	"errors"
	"reflect"
	"unsafe"
)

type Item[K PartKeyTypes] struct {
	PartKey K
	SortKey any
	Value   []byte
}

type ItemQuery[K PartKeyTypes] struct {
	PartKey K
	SortKey any
	Value   []byte

	// todo: put these into metadata? or we'll put them at api json lvl?
	FoundIn      FindStatus
	FoundSSLevel int8
	FoundSSIdx   uint32
	Deleted      bool
}

type DataType string

const (
	DTYPE_INT32   DataType = "int32"
	DTYPE_INT64   DataType = "int64"
	DTYPE_FLOAT32 DataType = "float32"
	DTYPE_FLOAT64 DataType = "float64"
	DTYPE_BYTES   DataType = "bytes"
	DTYPE_STRING  DataType = "string"
)

type FindStatus = uint8

const (
	FOUND_STATUS_UNKNOWN FindStatus = iota
	FOUND_AT_MEM
	FOUND_AT_BLOOM
	FOUND_AT_SS
	NOT_FOUND
)

func (q ItemQuery[K]) AsItem() *Item[K] {
	return &Item[K]{
		PartKey: q.PartKey,
		SortKey: q.SortKey,
		Value:   q.Value,
	}
}

type TypeInfo struct {
	Type          DataType
	IsDynamicSize bool
	StaticSize    uint32
}

func GetTypeInfo[T any]() (*TypeInfo, error) {
	typeInfo := &TypeInfo{
		IsDynamicSize: false,
	}
	var asd T

	switch any(asd).(type) {
	case int32:
		typeInfo.Type = DTYPE_INT32
	case int64:
		typeInfo.Type = DTYPE_INT64
	case float32:
		typeInfo.Type = DTYPE_FLOAT32
	case float64:
		typeInfo.Type = DTYPE_FLOAT64
	case string:
		typeInfo.Type = DTYPE_STRING
		typeInfo.IsDynamicSize = true
	case []byte:
		typeInfo.Type = DTYPE_BYTES
		typeInfo.IsDynamicSize = true
	default:
		// forbidden type
		return nil, errors.New("invalid data type")
	}

	if !typeInfo.IsDynamicSize {
		typeInfo.StaticSize = uint32(unsafe.Sizeof(asd))
	}

	return typeInfo, nil
}

func GetType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func GetSize(v interface{}) uint32 {
	switch v := any(v).(type) {
	case string:
		return uint32(len(v))
	case []byte:
		return uint32(len(v))
	default:
		return uint32(unsafe.Sizeof(v))
	}
}

func GetSizeTypeInfo(v interface{}, typeInfo *TypeInfo) uint32 {
	if !typeInfo.IsDynamicSize {
		return typeInfo.StaticSize
	}

	switch v := any(v).(type) {
	case string:
		return uint32(len(v))
	case []byte:
		return uint32(len(v))
	default:
		return uint32(unsafe.Sizeof(v))
	}
}
