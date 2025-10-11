package tests_system

import (
	"bytes"
	"iter"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

type TestIterable struct {
	NItems        uint32
	ItemSize      uint32
	RandomizeSeed uint64
	// FoundSSLevel  int8

	RndItem *core.Item[int32]
}

var rnd *rand.Rand

func (t *TestIterable) ItemIterator() iter.Seq[*core.Item[int32]] {
	if rnd == nil {
		panic("Pls call RandomizeTests()")
	}

	keys := make([]int32, 0, t.NItems)
	for range t.NItems {
		keys = append(keys, rnd.Int32N(1000000))
	}
	t.RndItem = &core.Item[int32]{
		PartKey: keys[rnd.IntN(int(t.NItems)-1)],
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return func(yield func(*core.Item[int32]) bool) {

		for i := range t.NItems {
			item := &core.Item[int32]{
				PartKey: keys[i],
				Value:   RandAsciiByte(int(t.ItemSize)),
			}

			// pick out a random item for later testing
			if t.RndItem.PartKey == item.PartKey {
				t.RndItem = item
			}

			if !yield(item) {
				return
			}
		}
	}
}

func (t *TestIterable) Len() uint32 {
	return uint32(t.NItems)
}

func (t *TestIterable) TotalKeySize() uint64 {
	return uint64(t.ItemSize * t.Len())
}

func (t *TestIterable) TotalValueSize() uint64 {
	return uint64(4 * t.Len())
}

func CapturePrint(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Close the writer and restore os.Stdout
	t.Cleanup(func() {
		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		t.Log(buf.String())
	})
}

func SetupTable(t *testing.T, memsize uint64, blocksize uint32, cleanup bool) *lsm.CfgTable {
	// Create a relative folder
	cwd, _ := os.Getwd()
	dbPath := filepath.Join(cwd, "..", "tmp")

	// Capture print
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// clean up capturing & restore os.Stdout
	t.Cleanup(func() {
		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		println(buf.String())

		// clean up files
		if cleanup {
			err := os.RemoveAll(dbPath)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	if memsize == 0 {
		memsize = 16348
	}

	if blocksize == 0 {
		blocksize = 65536
	}

	return &lsm.CfgTable{
		Name: "table1",
		MemTable: memtable.CfgMemtable{
			Type:        memtable.MEMTYPE_REDBLACK,
			MaxByteSize: memsize,
		},
		SSTable: sstable.CfgSSTable{
			DBPath:               dbPath,
			CompressionBlockSize: blocksize,
			MinIdxInterval:       128,
			MaxIdxInterval:       128,
			MaxSumSize:           5242880,
			BloomFilter: bloom.CfgBloomFilter{
				FalsePositiveRate: 0.1,
			},
		},
	}
}

func RandomizeTests(seed uint64) {
	if seed != 0 {
		rnd = rand.New(rand.NewPCG(seed, 2))
		return
	}

	f, err := os.Open("/dev/urandom")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var b [32]byte
	_, err = f.Read(b[:])
	if err != nil {
		panic(err)
	}

	// now := uint64(time.Now().UnixNano())
	rnd = rand.New(rand.NewChaCha8(b))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandAsciiByte(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rnd.Uint64(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rnd.Uint64(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}
