package commands

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core/ioutils"

	"github.com/pelletier/go-toml/v2"
)

const (
	TABLE_INFO CommandType = iota + 10
	CREATE_TABLE
	UPDATE_TABLE
	DROP_TABLE
	DESCRIBE_TABLE
	FLUSH_TABLE
	// SETCONFIG
	BALANCE_INDEX
	REPARTITION_TABLE
)

func GetTable(table lsm.LSMTreeTableInterface, conn net.Conn) error {
	// TODO: @later: make client request return format? for now only JSON is supported

	cfgContent, err := TableCfgBytes(table.TableInfo(), 1)

	if err != nil {
		return err
	}

	return SendCommandResponse(TABLE_INFO, conn, cfgContent)
}

func CreateTable(node Node, conn net.Conn) error {
	var cfgType uint8
	cfg, err := ReadTableCfgIO(conn, &cfgType)

	if err != nil {
		return err
	}

	if cfg.Name == "" {
		return SendCommandResponse(CREATE_TABLE, conn)
	}

	// @TODO: make this overriddable later?
	cfg.SSTable.DBPath = node.GetDBPath()

	cfg.ApplyDefaults()

	err = cfg.WriteToDisc()
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Created table at %s!", cfg.Name, cfg.SSTable.DBPath)

	node.LoadTableFromDisc(cfg.Name)

	err = binary.Write(conn, binary.BigEndian, CREATE_TABLE)
	if err != nil {
		return err
	}

	cfgContent, err := TableCfgBytes(cfg, cfgType)

	if err != nil {
		return err
	}

	return SendCommandResponse(CREATE_TABLE, conn, cfgContent)
}

func DropTable(table lsm.LSMTreeTableInterface, conn net.Conn) error {
	fmt.Println("DROP TABLE:", table.TableName())
	return nil
}

func SetConfigTable(table lsm.LSMTreeTableInterface, conn net.Conn) error {
	fmt.Println("EDIT TABLE:", table.TableName())

	return nil
}

func ListTables(node Node, conn net.Conn) error {
	tt := node.ListTables()
	var tables []lsm.CfgTable = make([]lsm.CfgTable, 0, len(tt))

	for _, table := range tt {
		tables = append(tables, *table.TableInfo())
	}

	tablesJson, err := json.Marshal(tables)

	if err != nil {
		return err
	}

	return SendCommandResponse(LIST_TABLES, conn, tablesJson)
}

func FlushTable(table lsm.LSMTreeTableInterface, conn net.Conn) error {
	err := table.FlushMemToDisc()

	if err != nil {
		return err
	}

	return SendCommandResponse(FLUSH_TABLE, conn)
}

func HandleTableCommand(table lsm.LSMTreeTableInterface, conn net.Conn, cmd CommandType) error {
	var err error

	// @TODO: $ITT: support json and toml too for table cfg?

	switch cmd {
	case TABLE_INFO:
		// return cfg
		err = GetTable(table, conn)
	// case UPD_TABLE:
	//   // @TODO: @later
	// 	SetConfigTable(table, conn)
	case DROP_TABLE:
		err = DropTable(table, conn)
	case FLUSH_TABLE:
		err = FlushTable(table, conn)
	default:
		return fmt.Errorf("invalid table command %d", cmd)
	}

	return err
}

func ReadTableCfgIO(reader io.Reader, cfgType *uint8) (*lsm.CfgTable, error) {
	err := binary.Read(reader, binary.BigEndian, cfgType)
	if err != nil {
		return nil, fmt.Errorf("parse type error: %s", err)
	}

	cfgContent, err := ioutils.ReadDynamic[uint32](reader)

	if err != nil {
		return nil, err
	}

	cfg := &lsm.CfgTable{}

	switch *cfgType {
	case 0: // toml
		err := toml.Unmarshal(cfgContent, cfg)

		if err != nil {
			return nil, fmt.Errorf("toml parse error: %s", err)
		}
	case 1: // json
		err := json.Unmarshal(cfgContent, cfg)

		if err != nil {
			return nil, fmt.Errorf("json parse error: %s", err)
		}
	default:
		return nil, fmt.Errorf("invalid cfg format: %d", cfgType)
	}

	return cfg, nil
}

func TableCfgBytes(cfg *lsm.CfgTable, cfgType uint8) ([]byte, error) {
	var cfgContent []byte
	var err error

	switch cfgType {
	case 0: // toml
		cfgContent, err = toml.Marshal(cfg)

		if err != nil {
			return nil, fmt.Errorf("toml write error: %s", err)
		}
	case 1: // json
		cfgContent, err = json.Marshal(cfg)

		if err != nil {
			return nil, fmt.Errorf("json write error: %s", err)
		}
	default:
		return nil, fmt.Errorf("invalid cfg format: %d", cfgType)
	}

	return cfgContent, nil
}
