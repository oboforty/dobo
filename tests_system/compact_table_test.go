package tests_system

import (
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

	iter := TestIterable{
		Randomize:    true,
		NItems:       n_items,
		ItemSize:     VAL_SIZE,
		FoundSSLevel: 0,
	}

	// Arrange - flush to disc
	if err := sstable1.WriteToDisc(&iter); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := sstable2.WriteToDisc(&iter); err != nil {
		t.Error(err)
		t.FailNow()
	}

	println("###################################################\n")
	// Act - compact them, should be merge sort across all data structures
	sstable3, err := sstable.CompactTables(sstable1, sstable2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	println(sstable3.GenerationId)
}
