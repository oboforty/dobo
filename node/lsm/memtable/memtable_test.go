package memtable

import (
	"math/rand"
	"testing"
	"time"
)

func TestIteration(t *testing.T) {
	rb, err := NewRedBlack[int64, string]()

	if err != nil {
		t.Error(err)
	}

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
