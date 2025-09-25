package tests_system

import (
	"path/filepath"
	"testing"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/sstable"
)

// Creates and saves an LSM table to disc,
// Then checks if the same table's configs can be reloaded from a fresh start
func TestWriteReadConfig(t *testing.T) {
	RandomizeTests(0)

	// Arrange - random cfg values
	cfg := SetupTable(t, 1234, 2*64*1024, true)
	pkt, _ := core.GetTypeInfo[int32]()

	sst1 := sstable.New[int32](&cfg.SSTable, cfg.Name, *pkt, 0)

	// Arrange - just write one item to have summary & index page
	if err := sst1.WriteToDisc(&TestIterable{
		RandomizeSeed: 1337,
		NItems:        2,
		ItemSize:      10,
		FoundSSLevel:  0,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Arrange - write config to disc
	cfg.KeyType = "int32"
	err := cfg.WriteToDisc()
	if err != nil {
		t.Fatal(err)
	}

	tablePath := filepath.Join(cfg.SSTable.DBPath, cfg.Name)
	treeI, err := lsm.NewFromDisc(tablePath)
	if err != nil {
		t.Fatal(err)
	}

	tree := treeI.(*lsm.LSMTreeTable[int32])
	if tree.TableName() != cfg.Name {
		t.Error("Invalid tablename")
		t.FailNow()
	}

	if tree.SSTables[0].GetGenerationId() != 0 {
		t.Error("Invalid GenerationId")
		t.FailNow()
	}

	// Assert - file content
	// @TODO....
}
