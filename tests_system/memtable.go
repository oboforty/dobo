package tests_system

import (
	"testing"

	"bytes"
	"encoding/json"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
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
	table.Upsert(&core.Item[int32]{
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
