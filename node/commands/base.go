package commands

type CommandType = byte

type CommandDescription struct {
	Name string

	// true for commands which require no LSM Table to be loaded
	NoTableNeeded bool

	// @TODO: add bool to load PartKey & Value from conn IO?
}

var CMD_DESCR = map[byte]CommandDescription{
	LIST_TABLES: {Name: "List Tables", NoTableNeeded: true},

	TABLE_INFO:   {Name: "Table Info"},
	CREATE_TABLE: {Name: "Create Table", NoTableNeeded: true},
	UPDATE_TABLE: {Name: "Update Table Settings"},
	DROP_TABLE:   {Name: "Drop Table"},

	GET_ITEM: {Name: "Get"},
	PUT_ITEM: {Name: "Put"},
	UPD_ITEM: {Name: "Update"},
	DEL_ITEM: {Name: "Delete"},

	QRY_ITEMS: {Name: "Query Items"},
	PUT_ITEMS: {Name: "Insert Items"},
	UPD_ITEMS: {Name: "Update Items"},
	DEL_ITEMS: {Name: "Delete Items"},
}
