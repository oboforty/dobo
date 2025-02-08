package node

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

type CommandType = byte

const (
	GET_TABLE CommandType = iota + 10
	PUT_TABLE
	UPD_TABLE
	DEL_TABLE
)
const (
	GET_ITEM CommandType = iota + 20
	PUT_ITEM
	UPD_ITEM
	DEL_ITEM
)
const (
	QRY_ITEMS CommandType = iota + 30
	PUT_ITEMS
	UPD_ITEMS
	DEL_ITEMS
)

func GetItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) {
	item := table.Get(*key)

	fmt.Println("GET:", item)
}

func PutItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P, value []byte) {
	item := core.ItemWrite[P]{
		PartKey: *key,
		Value:   value,
	}

	fmt.Println("PUT:", item.PartKey)
	table.Upsert(&item)
}

// func UpdateItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P, value []byte) {
// 	var val []byte
// 	ioutils.GetVal(&val)
// 	item := table.Update(key)
// 	fmt.Println("UPDATE:", item)
// }

// func DeleteItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) {
// 	item := table.Delete(key)
// 	fmt.Println("DELETE:", item)
// }

func (node *Node) handleCommands(conn net.Conn) {
	defer conn.Close()

	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Printf("[cmdver] cmd typ error: %s", err)
			continue
		}

		var cmd CommandType = b[0]
		var tableName string

		err = ioutils.ReadDynamicValue[uint8](conn, &tableName)
		if err != nil {
		}
		tableIF, ok := node.Tables[tableName]
		if !ok {
		}

		// Item commands
		if 20 <= cmd && cmd <= 39 {
			switch table := tableIF.(type) {
			case *lsm.LSMTreeTable[int32]:
				handleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[int64]:
				handleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[float32]:
				handleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[float64]:
				handleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[string]:
				handleItemCommand(table, conn, cmd)
			default:
				fmt.Println("Unsupported table type")
			}
		}
	}
}

func handleItemCommand[T cmp.Ordered](table *lsm.LSMTreeTable[T], conn net.Conn, cmd CommandType) error {
	var key T
	ioutils.ReadDynamicValue[uint32](conn, &key)
	var value []byte
	var err error

	switch cmd {
	case GET_ITEM:
		GetItem(table, conn, &key)
	case PUT_ITEM:
		value, err = ioutils.ReadDynamic[uint32](conn)
		PutItem(table, conn, &key, value)
	// case UPD_ITEM:
	// 	value, err = ioutils.ReadDynamic[uint32](conn)
	// 	UpdateItem(table, conn, &key, value)
	// case DEL_ITEM:
	// 	DeleteItem(table, conn, &key)
	default:
		fmt.Println("Invalid command")
	}

	return err
}
