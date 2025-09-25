package core

import (
	"bytes"
	"cmp"
	"time"
)

type PartKeyTypes interface {
	~int32 | ~int64 | ~float32 | ~float64 | ~string | ~[]byte |
		time.Time
}

// UberComparator returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
func UberComparator[T PartKeyTypes](x, y T) int {

	switch a := any(x).(type) {
	case int:
		b := any(y).(int)
		return cmp.Compare(a, b)
	case int32:
		b := any(y).(int32)
		return cmp.Compare(a, b)
	case int64:
		b := any(y).(int64)
		return cmp.Compare(a, b)
	case float32:
		b := any(y).(float32)
		return cmp.Compare(a, b)
	case float64:
		b := any(y).(float64)
		return cmp.Compare(a, b)
	case string:
		b := any(y).(string)
		return cmp.Compare(a, b)
	case []byte:
		b := any(y).([]byte)
		return bytes.Compare(a, b)
	case time.Time:
		b := any(y).(time.Time)

		switch {
		case a.After(b):
			return 1
		case a.Before(b):
			return -1
		default:
			return 0
		}
	}

	return 0
}
