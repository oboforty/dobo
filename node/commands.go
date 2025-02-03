package node

import (
	"io"

	cmd "github.com/oboforty/dobo/node/commands"
	ser "github.com/oboforty/dobo/node/serialize"
)

var cmds map[byte]ser.CommandConstructor = map[byte]ser.CommandConstructor{
	// Table
	10: func(r io.Reader) ser.Command { return &cmd.GetTableCmd{TableCmd: *ser.ParseTableCmd(r)} },
	11: func(r io.Reader) ser.Command { return &cmd.CreateTableCmd{TableCmd: *ser.ParseTableCmd(r)} },
	12: func(r io.Reader) ser.Command { return &cmd.EditTableSettingsCmd{TableCmd: *ser.ParseTableCmd(r)} },
	13: func(r io.Reader) ser.Command { return &cmd.DropTableCmd{TableCmd: *ser.ParseTableCmd(r)} },

	// Item
	20: func(r io.Reader) ser.Command { return &cmd.GetItemCmd{ItemCmd: *ser.ParseItemCmd(r)} },
	21: func(r io.Reader) ser.Command { return &cmd.PutItemCmd{ItemCmd: *ser.ParseItemCmd(r)} },
	22: func(r io.Reader) ser.Command { return &cmd.UpdateItemCmd{ItemCmd: *ser.ParseItemCmd(r)} },
	23: func(r io.Reader) ser.Command { return &cmd.DelItemCmd{ItemCmd: *ser.ParseItemCmd(r)} },

	// Bulk -- @TODO
	// Bulk - @TODO: slice of KVP struct?
	// 30: {Name: "query_items", Flags: CMDP_TABLE_ITEM},
	// 31: {Name: "upsert_items", Flags: CMDP_TABLE_ITEM},
	// 32: {Name: "update_items", Flags: CMDP_TABLE_ITEM},
	// 33: {Name: "remove_items", Flags: CMDP_TABLE_ITEM},
}
