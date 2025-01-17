package core

import (
	"errors"
	"unsafe"
)

type TypeInfo struct {
	IsDynamicSize bool
	StaticSize    uint

	InsertedSize uint
}

func GetTypeInfo[T any]() (*TypeInfo, error) {
	var asd T
	var size uint
	var isdynsize bool

	switch any(asd).(type) {
	// case int:
	// case int8:
	// case int16:
	case int32:
	case int64:
	// case uint:
	// case uint8:
	// case uint16:
	// case uint32:
	// case uint64:
	// case uintptr:
	case float32:
	case float64:
		// case rune:
		size = uint(unsafe.Sizeof(asd))
	case string:
	case []byte:
		isdynsize = true
	default:
		// forbidden type
		return nil, errors.New("invalid data type")
	}

	return &TypeInfo{
		StaticSize:    size,
		IsDynamicSize: isdynsize,
	}, nil
}

func GetSize[T any](v T) uint {
	switch v := any(v).(type) {
	case string:
	case []byte:
		return uint(len(v))
	}

	return 0
}
