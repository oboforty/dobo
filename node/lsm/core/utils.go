package core

import (
	"errors"
	"unsafe"
)

type TypeInfo struct {
	Type          DataType
	IsDynamicSize bool
	StaticSize    uint
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
		typeInfo.StaticSize = uint(unsafe.Sizeof(asd))
	}

	return typeInfo, nil
}

func GetSize(v interface{}) uint {
	switch v := any(v).(type) {
	case string:
		return uint(len(v))
	case []byte:
		return uint(len(v))
	default:
		return uint(unsafe.Sizeof(v))
	}
}

func GetSizeTypeInfo(v interface{}, typeInfo *TypeInfo) uint {
	if !typeInfo.IsDynamicSize {
		return typeInfo.StaticSize
	}

	switch v := any(v).(type) {
	case string:
		return uint(len(v))
	case []byte:
		return uint(len(v))
	default:
		return uint(unsafe.Sizeof(v))
	}
}
