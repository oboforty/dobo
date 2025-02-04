package tests_system

import (
	"testing"

	"bytes"
	"encoding/json"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

// Tests read & write on memory object only
func TestReadWriteMemTable(t *testing.T) {
	CapturePrint(t)

	table, err := lsm.New[int32](&lsm.CfgTable{
		MemTable: memtable.CfgMemtable{
			Type:        memtable.MEMTYPE_REDBLACK,
			MaxByteSize: 300,
		},
	})

	if err != nil {
		t.Error(err)
		t.FailNow()
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
		t.FailNow()
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
	if string(val) != `{"asd":213352,"nested":{"fos":"opoly"},"tes":"show"}` {
		t.Logf("Wrong string: %s", string(val))
		t.FailNow()
	}
}

// Tests writing to disc
func TestFlushMemTable(t *testing.T) {
	// Arrange: Memtable should be flushed after 10 items
	const N_ITEMS int32 = 10
	unitSize := memtable.RB_NODE_PTRS_SIZE + 4 + 8
	cfg := SetupTable(t, true,
		unitSize*uint32(N_ITEMS), // mem size
		64*1024,                  // compression block size
	)

	table, err := lsm.New[int32](cfg)
	if err != nil {
		t.Fatal(err)
	}

	// fill memtable up with string[ 8]
	for i := range N_ITEMS {
		table.Upsert(&core.ItemWrite[int32]{
			PartKey: i,
			Value:   RandAsciiByte(8),
		})
	}

	if !table.MemTable.IsFull() {
		expectedSize := unitSize * uint32(N_ITEMS)

		t.Error("Memtable was expected to be full after: ", expectedSize)
		t.FailNow()
	}

	// Act - flush memtable to disc
	if err = table.FlushMemToDisc(); err != nil {
		t.Error(err)
		t.FailNow()
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

// Writes a bunch of items (10 blocks) to disc
// Then tests if an item can be retrieved
func TestWriteReadItemSSTable(t *testing.T) {
	// Arrange: Memtable shou ld be flushed after 10 items
	const VAL_SIZE uint = 10
	const BLOCK_SIZE = 64 * 1024
	n_items := 10 * (BLOCK_SIZE / VAL_SIZE)
	cfg := SetupTable(t, true, 0, BLOCK_SIZE)
	pkt, _ := core.GetTypeInfo[int32]()

	sstable := sstable.New[int32](
		&cfg.SSTable, cfg.Name, *pkt, 0,
	)

	// at seed=1337, generated item values are:
	// 14,20,31,33,74,89,96,259
	iter := TestIterable{
		Randomize:    false,
		NItems:       n_items,
		ItemSize:     VAL_SIZE,
		FoundSSLevel: 0,
	}

	// Act - write to disc
	if err := sstable.WriteToDisc(&iter); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Assert - correct idx file
	// idx file entries should be (3 * 4 + 4 = 16 bytes (key itself is ))
	expectedItem := iter.RndItem

	if expectedItem == nil {
		t.Fatal("no expectedItem.. why?")
	}

	// Act - read from disc
	actualItem := sstable.Get(expectedItem.PartKey)
	// actualItem := sstable.Get(89)

	if actualItem == nil || actualItem.Value == nil {
		t.Error("Item not found")
		t.FailNow()
	}

	if expectedItem.PartKey != actualItem.PartKey || !bytes.Equal(expectedItem.Value, actualItem.Value) {
		t.Error("Malformed item found")
		t.FailNow()
	}

	// @TODO: assert dat & idx file content?
}

// Creates and saves an LSM table to disc,
// Then checks if the same table's configs can be reloaded from a fresh start
func TestWriteReadConfig(t *testing.T) {
	// Arrange - random cfg values
	cfg := SetupTable(t, true, 1234, 64*1024)
	pkt, _ := core.GetTypeInfo[int32]()

	sst1 := sstable.New[int32](&cfg.SSTable, cfg.Name, *pkt, 0)

	// Arrange - just write one item to have summary & index page
	if err := sst1.WriteToDisc(&TestIterable{
		Randomize: false, NItems: 2, ItemSize: 10, FoundSSLevel: 0,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// sst1.(&core.ItemWrite[int32]{
	// 	PartKey: 12345,
	// 	Value:   []byte{0, 1, 2, 3, 4},
	// })

	// ss.WriteToDisc(memtOld)

	// err = table1.FlushMemToDisc()
	// if err = table.FlushMemToDisc(); err != nil {
	// 	t.Error(err)
	// 	t.FailNow()
	// }

}
