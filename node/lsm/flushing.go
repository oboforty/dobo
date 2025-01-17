package lsm

import "cmp"

func FlushMemTable[P cmp.Ordered, V any](memt MemTable[P, V]) (SSTable[P, V], error) {

	// todo: iter, sort, write sorted to disk
	// todo: write idx (&summary @later) file
	// todo: collect stats & write? @later

	// todo: detect HERE? or in a scheduled task when to trigger the compaction goroutine?

	// *SSTable[P, V]
	return nil, nil
}
