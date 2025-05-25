package tests_system

import (
	"bytes"
	"testing"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable"
)

func TestTablesCompaction(t *testing.T) {
	// Arrange: two SS Tables
	const VAL_SIZE uint32 = 10
	const BLOCK_SIZE = 1024
	const n_items = 2 * (uint32(BLOCK_SIZE) / VAL_SIZE)
	cfg := SetupTable(t, 0, BLOCK_SIZE, false)
	pkt, _ := core.GetTypeInfo[int32]()

	sstable1 := sstable.New[int32](
		&cfg.SSTable, cfg.Name, *pkt, 0,
	)

	sstable2 := sstable.New[int32](
		&cfg.SSTable, cfg.Name, *pkt, 1,
	)

	iter1 := TestIterable{
		RandomizeSeed: 0,
		NItems:        n_items,
		ItemSize:      VAL_SIZE,
		FoundSSLevel:  0,
	}

	iter2 := TestIterable{
		RandomizeSeed: 0,
		NItems:        n_items,
		ItemSize:      VAL_SIZE,
		FoundSSLevel:  0,
	}

	// Arrange - flush to disc
	if err := sstable1.WriteToDisc(&iter1); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := sstable2.WriteToDisc(&iter2); err != nil {
		t.Error(err)
		t.FailNow()
	}

	expectedItem1 := iter1.RndItem
	expectedItem2 := iter2.RndItem

	// Act - compact them, should be merge sort across all data structures
	sstable3, err := sstable.CompactTables(sstable1, sstable2, 2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	sstable1.Get(expectedItem1.PartKey)
	return

	// Assert - one item from both tables can be found
	actualItem := sstable3.Get(expectedItem1.PartKey)

	if actualItem == nil || actualItem.Value == nil {
		t.Error("Item not found #1")
		t.FailNow()
	}

	if !bytes.Equal(expectedItem1.Value, actualItem.Value) {
		t.Log("Expected value: ", expectedItem1.Value)
		t.Log("Actual value:   ", actualItem.Value)

		t.Errorf("Malformed item found #1")
		t.FailNow()
	}

	if expectedItem1.PartKey != actualItem.PartKey {
		t.Log("Expected key: ", expectedItem1.PartKey)
		t.Log("Actual key:   ", actualItem.PartKey)

		t.Errorf("Malformed item found #1")
		t.FailNow()
	}

	actualItem = sstable3.Get(expectedItem2.PartKey)

	if actualItem == nil || actualItem.Value == nil {
		t.Error("Item not found #2")
		t.FailNow()
	}

	if !bytes.Equal(expectedItem2.Value, actualItem.Value) {
		t.Log("Expected value: ", expectedItem2.Value)
		t.Log("Actual value:   ", actualItem.Value)

		t.Errorf("Malformed item found #2")
		t.FailNow()
	}

	if expectedItem2.PartKey != actualItem.PartKey {
		t.Log("Expected key: ", expectedItem2.PartKey)
		t.Log("Actual key:   ", actualItem.PartKey)

		t.Errorf("Malformed item found #2")
		t.FailNow()
	}

	// println(sstable3.GenerationId)
}
