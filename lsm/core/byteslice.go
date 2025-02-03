package core

import "bytes"

type ByteSlice struct {
	data []byte
}

// Compare implements comparison for ByteSlice
func (b ByteSlice) Compare(other ByteSlice) int {
	return bytes.Compare(b.data, other.data)
}
