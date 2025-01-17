package tests_system

import (
	"testing"

	"bytes"
	"encoding/json"

	"dobo/lsm"
	"dobo/lsm/core"
)

func TestReadWRite(t *testing.T) {
	table, err := lsm.New[int32, []byte](&lsm.CfgTable{})

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
	table.Upsert(&core.ItemWrite[int32, []byte]{
		PartKey: key,
		Value:   serialized,
	})

	item := table.Get(key)

	if item == nil || item.Value == nil {
		t.FailNow()
	}

	if !bytes.Equal(item.Value, serialized) {
		t.FailNow()
	}
}

func TestReadFromSSTable(t *testing.T) {
	table, err := lsm.New[int32, []byte](&lsm.CfgTable{})

	if err != nil {
		t.Error(err)
	}

	var key int32 = 1234567890
	item := table.Get(key)

	if item == nil || item.Value == nil {
		t.FailNow()
	}

	if item.Value == nil {
		t.Fail()
	}
}
