package tests_system

import (
	"bytes"
	"iter"
	rand2 "math/rand"
	rand "math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

type TestIterable struct {
	NItems       uint
	ItemSize     uint
	Randomize    bool
	FoundSSLevel int8

	RndItem *core.ItemQuery[int32]
}

func (t *TestIterable) ItemIterator() iter.Seq[*core.ItemQuery[int32]] {
	var rnd *rand.Rand
	if t.Randomize {
		rnd = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	} else {
		rnd = rand.New(rand.NewPCG(1337, 0))
	}

	keys := make([]int32, 0, t.NItems)
	for range t.NItems {
		keys = append(keys, rnd.Int32N(1000000))
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return func(yield func(*core.ItemQuery[int32]) bool) {

		for i := range t.NItems {
			item := &core.ItemQuery[int32]{
				PartKey: keys[i],
				Value:   RandAsciiByte(int(t.ItemSize)),

				FoundIn:      core.FOUND_AT_SS,
				FoundSSLevel: t.FoundSSLevel,
			}

			// pick out a random item for later testing
			if t.RndItem == nil && i > t.NItems/3 && rnd.Float32() > 0.9 {
				t.RndItem = item
			}

			if !yield(item) {
				return
			}
		}
	}
}

func (t *TestIterable) Size() uint32 {
	return uint32(t.NItems)
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

func SetupTable(t *testing.T, ensureTable bool, memsize uint32, blocksize uint32) *lsm.CfgTable {
	// Create a relative folder
	cwd, _ := os.Getwd()
	dbPath := filepath.Join(cwd, "..", "tmp")

	if ensureTable {
		if err := EnsurePath(dbPath); err != nil {
			panic(err)
		}
	}

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
		if ensureTable {
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
			BloomFilter: bloom.CfgBloomFilter{
				FalsePositiveRate: 0.1,
			},
		},
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand2.NewSource(time.Now().UnixNano())

func RandAsciiByte(n int) []byte {
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

	return b
}

func EnsurePath(tablePath string) error {
	if _, err := os.Stat(tablePath); err != nil {
		err = os.MkdirAll(tablePath, os.ModePerm)

		if err != nil {
			return os.ErrNotExist
		}
	}

	return nil
}
