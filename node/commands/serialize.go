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

type Node interface {
	Table(string) lsm.LSMTreeTableInterface
}

type Command interface {
	Run(Node, net.Conn) error
}

func GetTable[P cmp.Ordered](node Node, tableName string) (*lsm.LSMTreeTable[P], error) {
	t := node.Table(tableName)

	if t == nil {
		return nil, fmt.Errorf("not found table %s", tableName)
	}

	t2, ok := t.(*lsm.LSMTreeTable[P])

	if !ok {
		var v P
		info := core.GetType(v)
		return nil, fmt.Errorf("type mismatch %s is not of type %s", tableName, info)
	}

	return t2, nil
}

type CommandConstructor func(io.Reader) Command

func ParseItemCmd(reader io.Reader) *ItemCmd {
	cmdi := &ItemCmd{}
	err := ioutils.ReadDynamicValue[uint8](reader, &cmdi.Table)
	if err != nil {
		return nil
	}

	cmdi.PartKey, err = ioutils.ReadDynamic[uint32](reader)
	if err != nil {
		return nil
	}

	cmdi.Value, err = ioutils.ReadDynamic[uint32](reader)
	if err != nil {
		return nil
	}

	return cmdi
}

func ParseTableCmd(reader io.Reader) *TableCmd {
	cmdi := &TableCmd{}
	err := ioutils.ReadDynamicValue[uint8](reader, &cmdi.Table)
	if err != nil {
		return nil
	}

	return cmdi
}
