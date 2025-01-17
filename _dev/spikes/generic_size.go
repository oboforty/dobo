package main

import (
	"cmp"
	"unsafe"
)

type ItemWrite[P cmp.Ordered, V any] struct {
	PartKey P
	Value   V
}

func fos[T any](i []T) {

	println(len(i) * int(unsafe.Sizeof(i[0])))

}

func fos2[P cmp.Ordered, V any](item *ItemWrite[P, V]) {

	println(len(item.PartKey) * int(unsafe.Sizeof(item.PartKey)))

}

func main() {

	fos([]byte{32, 24, 123, 12, 1, 2, 4})

}
