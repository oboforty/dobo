package commands

import (
	"fmt"
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

func GetItem[P core.PartKeyTypes](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) error {
	item := table.Get(*key)

	// content, err := SerializeItem(item.AsItem(), 0)
	// if err != nil {
	// 	fmt.Printf("[%s] write error: %s (Get Item)", table.TableName(), err)
	// }
	if item == nil {
		// @TODO: handle error
		return nil
	}

	return SendCommand(GET_ITEM, conn, item.Value)
}

func PutItem[P core.PartKeyTypes](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) error {
	value, err := ioutils.ReadDynamic[uint32](conn)

	if err != nil {
		return err
	}

	item := &core.Item[P]{
		PartKey: *key,
		Value:   value,
	}

	table.Upsert(item)
	// content, err := SerializeItem(item, 0)

	// if err != nil {
	// 	return err
	// }

	return SendCommand(PUT_ITEM, conn)
}

// func UpdateItem[P core.PartKeyTypes](table *lsm.LSMTreeTable[P], conn net.Conn, key *P, value []byte) {
// 	var val []byte
// 	ioutils.GetVal(&val)
// 	item := table.Update(key)
// 	fmt.Println("UPDATE:", item)
// }

func DeleteItem[P core.PartKeyTypes](table *lsm.LSMTreeTable[P], conn net.Conn, key *P) {
	table.Delete(*key)
	fmt.Println("DELETE:", *key)
}

func HandleItemCommand[T core.PartKeyTypes](table *lsm.LSMTreeTable[T], conn net.Conn, cmd CommandType) error {
	var key T
	var err error

	err = ioutils.ReadDynamicValue[uint32](conn, &key)

	if err != nil {
		return err
	}

	switch cmd {
	case GET_ITEM:
		err = GetItem(table, conn, &key)
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

// func SerializeItem[P core.PartKeyTypes](item *core.Item[P], serType uint8) ([]byte, error) {
// 	var content []byte

// 	switch serType {
// 	case 0: // raw
// 		content = item.Value
// 	case 1: // json
// 	default:
// 		return nil, fmt.Errorf("invalid serialization format: %d", serType)
// 	}

// 	return content, nil
// }
