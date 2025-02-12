package commands

import (
	"cmp"
	"fmt"
	"io"
	"net"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

const (
	GET_ITEM CommandType = iota + 20
	PUT_ITEM
	UPD_ITEM
	DEL_ITEM
)

func GetItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) {
	item := table.Get(*key)

	err := WriteItemIO(conn, item, 0)

	if err != nil {
		fmt.Printf("[%s] write error: %s (Get Item)", table.TableName(), err)
	}
}

func PutItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) error {
	value, err := ioutils.ReadDynamic[uint32](conn)

	if err != nil {
		return err
	}

	item := core.ItemWrite[P]{
		PartKey: *key,
		Value:   value,
	}

	table.Upsert(&item)

	return nil
}

// func UpdateItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P, value []byte) {
// 	var val []byte
// 	ioutils.GetVal(&val)
// 	item := table.Update(key)
// 	fmt.Println("UPDATE:", item)
// }

func DeleteItem[P cmp.Ordered](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) {
	table.Delete(*key)
	fmt.Println("DELETE:", *key)
}

func HandleItemCommand[T cmp.Ordered](table *lsm.LSMTreeTable[T], conn net.Conn, cmd CommandType) error {
	var key T
	var err error

	err = ioutils.ReadDynamicValue[uint32](conn, &key)

	if err != nil {
		return err
	}

	switch cmd {
	case GET_ITEM:
		GetItem(table, conn, &key)
	case PUT_ITEM:
		err = PutItem(table, conn, &key)
	// case UPD_ITEM:
	// 	value, err = ioutils.ReadDynamic[uint32](conn)
	// 	UpdateItem(table, conn, &key, value)
	case DEL_ITEM:
		DeleteItem(table, conn, &key)
	default:
		return fmt.Errorf("invalid item command %d", cmd)
	}

	return err
}

func WriteItemIO[P cmp.Ordered](writer io.Writer, item *core.ItemQuery[P], serType uint8) error {
	var content []byte
	var err error

	switch serType {
	case 0: // raw
		content = item.Value
	case 1: // json
	default:
		return fmt.Errorf("invalid serialization format: %d", serType)
	}

	err = ioutils.WriteDynamic[uint32](writer, content)

	if err != nil {
		return err
	}

	return nil
}
