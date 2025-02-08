package commands

import (
	"net"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

type ItemCmd struct {
	Table   string
	PartKey []byte
	Value   []byte
}

type GetItemCmd struct {
	ItemCmd
}

func (p GetItemCmd) Run(node Node, conn net.Conn) error {

	return nil
}

type PutItemCmd struct {
	ItemCmd
}

func (p PutItemCmd) Run(node Node, conn net.Conn) error {
	tun := node.Table(p.Table)

	var err error
	var key interface{}

	switch table := tun.(type) {
	case *lsm.LSMTreeTable[int32]:
		var key int32
		ioutils.ReadDynamicValue[uint32](file, &key)

		// table, err := GetTable[int32](node, p.Table)
		// key := table.ConvertKey(p.PartKey)
		item := table.Get(key.(int32))

	}

	err

	var cmd string = "get"

	switch cmd {
	case "get":
	case "put":
	case "delete":
	case "get-bulk":
	case "put-bulk":
	case "delete-bulk":
	}

	if err != nil {
		return err
	}

	// table := node.Table(p.Table)

	// @TODI: $ITT: refactor -- rely on [P]

	// table.Upsert(node)

	return nil
}

type UpdateItemCmd struct {
	ItemCmd
}

func (p UpdateItemCmd) Run(node Node, conn net.Conn) error {

	return nil
}

type DelItemCmd struct {
	ItemCmd
}

func (p DelItemCmd) Run(node Node, conn net.Conn) error {

	return nil
}
