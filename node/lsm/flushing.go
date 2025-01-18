package lsm

import (
	"cmp"
)

func FlushMemTable[P cmp.Ordered, V any](memt MemTable[P, V]) (SSTable[P, V], error) {
	// tree is iterated in partition key order!
	for node := range memt.ItemIterator() {

		print(node.PartKey)
	}

	// todo: iter, sort, write sorted to disk
	// todo: write idx (&summary @later) file
	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	// *SSTable[P, V]
	return nil, nil
}
