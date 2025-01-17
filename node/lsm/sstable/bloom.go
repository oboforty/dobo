package sstable

import (
	"encoding/binary"
	"math"

	"github.com/bits-and-blooms/bloom/v3"
)

func Int32BFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	i := val.(uint32)
	nl := make([]byte, 4)
	binary.BigEndian.PutUint32(nl, i)

	return bloom.Test(nl)
}

func Int64BFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	i := val.(uint64)
	nl := make([]byte, 8)
	binary.BigEndian.PutUint64(nl, i)

	return bloom.Test(nl)
}

func Float32BFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	i := val.(float32)
	nl := make([]byte, 4)
	binary.BigEndian.PutUint32(nl, math.Float32bits(i))

	return bloom.Test(nl)
}

func Float64BFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	i := val.(float64)
	nl := make([]byte, 8)
	binary.BigEndian.PutUint64(nl, math.Float64bits(i))

	return bloom.Test(nl)
}

func BytesBFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	return bloom.Test(val.([]byte))
}

func StringBFCmp(bloom *bloom.BloomFilter, val interface{}) bool {
	return bloom.TestString(val.(string))
}

// @TODO: time
