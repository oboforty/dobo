package commands

import (
	"net"

	"github.com/oboforty/dobo/lsm"
)

const (
	CMD_OK CommandType = iota + 1
	_RESERVED
	NODE_INFO
	LIST_TABLES

	// @TODO: discover nodes, share info, rebalance token ring,
	// REPLICATE-TABLE,
	// CREATE-PARTITION, DELETE-PARTITION,

)

type Node interface {
	ListTables() map[string]lsm.LSMTreeTableInterface
	GetDBPath() string
	// Table(string) lsm.LSMTreeTableInterface
	LoadTableFromDisc(string)
}

func HandleNodeCommand(node Node, conn net.Conn, cmd CommandType) error {
	var err error

	switch cmd {
	case LIST_TABLES:
		err = ListTables(node, conn)
	case CREATE_TABLE:
		err = CreateTable(node, conn)
	}

	return err
}
