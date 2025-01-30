package serialize

import (
	"io"

	"github.com/oboforty/dobo/lsm/sstable/ioutils"
	"github.com/oboforty/dobo/node/commands"
)

type Command interface {
	Run() error
}

type CommandConstructor func(io.Reader) Command

func ParseItemCmd(reader io.Reader) *commands.ItemCmd {
	cmdi := &commands.ItemCmd{}
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

func ParseTableCmd(reader io.Reader) *commands.TableCmd {
	cmdi := &commands.TableCmd{}
	err := ioutils.ReadDynamicValue[uint8](reader, &cmdi.Table)
	if err != nil {
		return nil
	}

	return cmdi
}
