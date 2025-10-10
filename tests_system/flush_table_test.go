package tests_system

import (
	"testing"
	"time"

	"bytes"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

// Tests writing to disc
func TestFlushMemTable(t *testing.T) {
	RandomizeTests(0)

	// Arrange: SS Table
	const N_ITEMS int32 = 10
	unitSize := memtable.RB_NODE_PTRS_SIZE + 4 + 8
	cfg := SetupTable(t,
		unitSize*uint64(N_ITEMS), // mem size
		64*1024,                  // compression block size
		true,
	)

	table, err := lsm.New[int32](cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Arrange - fill memtable up with string[ 8]
	for i := range N_ITEMS {
		table.Upsert(&core.Item[int32]{
			PartKey: i,
			Value:   RandAsciiByte(8),
		})
	}

	// Assert - At this point memtable should be full
	if !table.MemTable.IsFull() {
		expectedSize := unitSize * uint64(N_ITEMS)

		t.Error("Memtable was expected to be full after: ", expectedSize)
		t.FailNow()
	}

	// Act - delete an item
	table.Delete(5)

	// Act - flush memtable to disc
	var ok bool
	startTime := time.Now()
	if ok, err = table.FlushMemToDisc(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !ok {
		t.Error("Flushed empty table!")
	}
	elapsed := time.Since(startTime)
	t.Logf("FLUSH took %d μs", elapsed.Microseconds())

	// Assert - memtable is cleared
	if table.MemTable.ByteSize() != 0 {
		t.Error("MemTable should haven be empty!")
		t.FailNow()
	}

	// Assert - get item from disc
	startTime = time.Now()
	item := table.Get(4)
	if item.PartKey != 4 || len(item.Value) != 8 {
		t.Error("SSTable GET fail")
		t.FailNow()
	}
	elapsed = time.Since(startTime)
	t.Logf("GET took %d μs", elapsed.Microseconds())

	// Assert - get deleted item from disc
	startTime = time.Now()
	item = table.Get(5)
	if item.PartKey != 5 || item.Value != nil || !item.Deleted {
		t.Error("SSTable GET deleted fail")
		t.FailNow()
	}
	elapsed = time.Since(startTime)
	t.Logf("GET deleted took %d μs", elapsed.Microseconds())

	// Assert - SSTable is created
	// if len(table.SSTables) != 1 || table.SSTables[0].GetGenerationId() != table.CurrentGenerationId() {
	// 	t.Error("SSTable GenId mismatch")
	// 	t.FailNow()
	// }

	// Act - reload table
	table2, err := lsm.New[int32](cfg)
	// Asssert - should be able to load bloom filter & summary
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	if len(table2.SSTables) != len(table.SSTables) {
		t.Errorf("Mismatch of SSTables: %d != %d", len(table.SSTables), len(table2.SSTables))
		t.FailNow()
	}
	// Assert - item can be still found, even if it's not in the MemTable
	item2 := table2.Get(4)
	if item2 == nil || item2.PartKey != 4 || len(item2.Value) != 8 {
		t.Error("SSTable GET #2 fail")
		t.FailNow()
	}
}

// Writes a bunch of items (10 blocks) to disc
// Then tests if an item can be retrieved
func TestWriteReadItemSSTable(t *testing.T) {
	// Arrange: Memtable shou ld be flushed after 10 items
	const VAL_SIZE uint32 = 10
	const BLOCK_SIZE = 5 * 1024
	const n_items = 10 * (uint32(BLOCK_SIZE) / VAL_SIZE)

	RandomizeTests(0)

	cfg := SetupTable(
		t,
		0,
		BLOCK_SIZE,
		true,
	)
	pkt, _ := core.GetTypeInfo[int32]()

	sstable := sstable.New[int32](
		&cfg.SSTable, cfg.Name, *pkt, 0,
	)

	// at seed=1337, generated item values are:
	// 14,20,31,33,74,89,96,259
	iter := TestIterable{
		RandomizeSeed: 1338,
		NItems:        n_items,
		ItemSize:      VAL_SIZE,
		FoundSSLevel:  0,
	}

	// Act - write to disc
	if err := sstable.WriteToDisc(&iter); err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("Random item key: %d value: %s", iter.RndItem.PartKey, string(iter.RndItem.Value))

	// Assert - correct idx file
	// idx file entries should be (3 * 4 + 4 = 16 bytes (key itself is ))
	expectedItem := iter.RndItem

	if expectedItem == nil {
		t.Fatal("no expectedItem.. why?")
	}

	// Act - read from disc
	startTime := time.Now()
	actualItem := sstable.Get(expectedItem.PartKey)
	elapsed := time.Since(startTime)
	t.Logf("GET %d took %d μs", expectedItem.PartKey, elapsed.Microseconds())

	if actualItem == nil || actualItem.Value == nil {
		t.Error("Item not found")
		t.FailNow()
	}

	if !bytes.Equal(expectedItem.Value, actualItem.Value) {
		t.Log("Expected value: ", expectedItem.Value)
		t.Log("Actual value:   ", actualItem.Value)

		t.Errorf("Malformed item found")
		t.FailNow()
	}

	if expectedItem.PartKey != actualItem.PartKey {
		t.Log("Expected key: ", expectedItem.PartKey)
		t.Log("Actual key:   ", actualItem.PartKey)

		t.Errorf("Malformed item found")
		t.FailNow()
	}
	// @TODO: assert dat & idx file content?
}
