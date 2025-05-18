package bloom_test

import (
	"testing"

	"github.com/oboforty/dobo/lsm/bloom"
)

func TestBloomFilter(t *testing.T) {

	bloom := bloom.New(bloom.CfgBloomFilter{
		FalsePositiveRate: 0.1,
		MaxItems:          65530,
	})

	var bb []byte = []byte{0, 2, 3}
	bloom.Add(bb)
}
