package tests_system

import (
	"testing"
	"unsafe"

	"bytes"
	"encoding/json"

	"dobo/lsm"
	"dobo/lsm/core"
	"dobo/lsm/memtable"
	"dobo/lsm/sstable"
	"dobo/lsm/utils"
)

func TestReadWRite(t *testing.T) {
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
	var key int32 = 1234567890
	const N_ITEMS int32 = 10

	unitSize := memtable.RB_NODE_PTRS_SIZE + uint(unsafe.Sizeof(key)) + 8
	table, err := lsm.New[int32](&lsm.CfgTable{
		Name: "test",

		MemTable: memtable.CfgMemtable{
			Type:        memtable.MEMTYPE_REDBLACK,
			MaxByteSize: unitSize * uint(N_ITEMS),
		},
		SSTable: sstable.CfgSSTable{
			// @TODO: conver from relative to tests into absolute path
			BasePath:                  "/home/rajmund_csombordi/dev/dobo/",
			DynamicValueSerialization: "jsonb",
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

	// Act
	if err = table.FlushMemToDisc(); err != nil {
		t.Error(err)
	}

	// Assert - memtable is cleared
	if table.MemTable.ByteSize() != 0 {
		t.FailNow()
	}

	// Assert - SSTable is created
	if len(table.SSTables) != 1 {
		t.FailNow()
	}

}

// func TestReadFromSSTable(t *testing.T) {
// 	table, err := lsm.New[int32, []byte](&lsm.CfgTable{})

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	var key int32 = 1234567890
// 	item := table.Get(key)

// 	if item == nil || item.Value == nil {
// 		t.FailNow()
// 	}

// 	if item.Value == nil {
// 		t.Fail()
// 	}
// }
