package memtable

import (
	"math/rand"
	"testing"
	"time"
)

func TestIteration(t *testing.T) {
	rb := NewRedBlack[int64]()

	iters := 100000
	for i := range iters {
		key := rand.Int63()

		keylen := int(rand.Int31())
		if keylen > 4000 {
			if i > 4000 {
				keylen = 2000
			} else {
				keylen = i
			}
		}

		str := RandStringBytesMaskImprSrc(keylen)
		rb.tree.Put(key, str)
	}

	start := time.Now()

	var prevKey int64 = -1
	for node := range rb.ItemIterator() {
		if node.PartKey < prevKey {
			t.Log("Incorrect key order in iteration!")
			t.FailNow()
		}

		prevKey = node.PartKey
	}
	t.Logf("Tree size: %d", rb.tree.Size())
	t.Logf("Full Iteration took %s", time.Since(start))
}
