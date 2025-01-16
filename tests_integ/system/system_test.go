package tests_system

import (
	"dobo/tables"
	"dobo/tables/core"
	"testing"
)

func TestBuildTree(t *testing.T) {

	table := &tables.LSMTable{}

	table.Init(core.DTYPE_INT32)

}
