package tests_system

import (
	"iter"
	"sort"
	"testing"
	"time"
	"unsafe"

	"bytes"
	"encoding/json"

	"dobo/lsm"
	"dobo/lsm/core"
	"dobo/lsm/memtable"
	"dobo/lsm/sstable"
	"dobo/lsm/utils"

	"math/rand/v2"
)

func TestReadWriteMemTable(t *testing.T) {
	table, err := lsm.New[int32](&lsm.CfgTable{
		MemTable: memtable.CfgMemtable{
			Type:        memtable.MEMTYPE_REDBLACK,
			MaxByteSize: 300,
		},
	})

	if err != nil {
		t.Error(err)
	}

	data := map[string]interface{}{
		"tes": "show",
		"asd": 213352,
		"nested": map[string]interface{}{
			"fos": "opoly",
		},
	}

	serialized, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	var key int32 = 1234567890
	table.Upsert(&core.ItemWrite[int32]{
		PartKey: key,
		Value:   serialized,
	})

	item := table.Get(key)

	if item == nil || item.Value == nil {
		t.FailNow()
	}

	val := item.Value
	if !bytes.Equal(val, serialized) {
		t.FailNow()
	}
}

func TestFlushMemTable(t *testing.T) {
	CapturePrint(t)

	// Arrange: Memtable shou ld be flushed after 10 items
	const N_ITEMS int32 = 10

	unitSize := memtable.RB_NODE_PTRS_SIZE + uint(unsafe.Sizeof(0)) + 8
	table, err := lsm.New[int32](&lsm.CfgTable{
		Name: "test",

		MemTable: memtable.CfgMemtable{
			Type:        memtable.MEMTYPE_REDBLACK,
			MaxByteSize: unitSize * uint(N_ITEMS),
		},
		SSTable: sstable.CfgSSTable{
			// @TODO: conver from relative to tests into absolute path
			BasePath: "/home/rajmund_csombordi/dev/dobo/nemtom",
			// DataSerialization: "jsonb",
			CompressionBlockSize: 64 * 1024,
		},
	})

	if err != nil {
		t.Error(err)
	}

	// fill memtable up with string[ 8]
	for i := range N_ITEMS {
		table.Upsert(&core.ItemWrite[int32]{
			PartKey: i,
			Value:   utils.RandAsciiByte(8),
		})
	}

	if !table.MemTable.IsFull() {
		t.FailNow()
	}

	// Act - flush memtable to disc
	if err = table.FlushMemToDisc(); err != nil {
		t.Error(err)
	}

	// Assert - memtable is cleared
	if table.MemTable.ByteSize() != 0 {
		t.FailNow()
	}

	// Assert - SSTable is created
	if len(table.SSTables) != 1 && table.SSTables[0].GetGenerationId() == table.CurrentGenerationId() {
		t.FailNow()
	}
}

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
				Value:   utils.RandAsciiByte(int(t.ItemSize)),

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

func TestWriteReadSSTable(t *testing.T) {
	CapturePrint(t)

	// Arrange: Memtable shou ld be flushed after 10 items
	const VAL_SIZE uint = 10
	const BLOCK_SIZE = 64 * 1024
	n_items := 10 * (BLOCK_SIZE / VAL_SIZE)

	dbpath := "/home/rajmund_csombordi/dev/dobo/"
	tablename := "nemtom"

	pkt, _ := core.GetTypeInfo[int32]()

	sstable := sstable.New[int32](
		&sstable.CfgSSTable{
			// @TODO: conver from relative to tests into absolute path
			BasePath:             dbpath,
			CompressionBlockSize: BLOCK_SIZE,
			// DataSerialization: "jsonb",
		}, tablename, *pkt, 0,
	)

	iter := TestIterable{
		// at seed=1337, item values are 14,20,31,33,74,89,96,259
		Randomize:    false,
		NItems:       n_items,
		ItemSize:     VAL_SIZE,
		FoundSSLevel: 0,
	}

	// Act - write to disc
	err := sstable.WriteToDisc(&iter)

	if err != nil {
		t.Error(err)
	}

	// Assert - correct idx file
	// idx file entries should be (3 * 4 + 4 = 16 bytes (key itself is ))

	// Act - read from disc
	expectedItem := iter.RndItem
	actualItem := sstable.Get(expectedItem.PartKey)
	// actualItem := sstable.Get(89)

	if actualItem == nil || actualItem.Value == nil {
		t.Error("Item not found")
		t.FailNow()
	}

	if expectedItem.PartKey != actualItem.PartKey || bytes.Equal(expectedItem.Value, actualItem.Value) {
		t.Error("Malformed item found")
		t.FailNow()
	}

	// assert that there are 10 blocks
}
