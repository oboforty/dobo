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
	// @TODO:
	// SETCONFIG,
	// CREATE-TABLE, DELETE-TABLE,
	// BALANCE-INDEX,
	// REPARTITION-TABLE
)

func GetTable(table lsm.LSMTreeTableInterface, conn net.Conn) error {
	fmt.Println("GET TABLE:", table.TableName())
	return nil
}

func CreateTable(node Node, conn net.Conn) error {
	var cfgType uint8
	cfg, err := ReadTableCfgIO(conn, &cfgType)

	if err != nil {
		return err
	}

	if cfg.Name == "" {
		binary.Write(conn, binary.BigEndian, CMD_ERROR)
		binary.Write(conn, binary.BigEndian, uint8(1))
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

	err = WriteTableCfgIO(conn, cfg, cfgType)

	if err != nil {
		return err
	}

	return nil
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
	var tables []string = make([]string, len(tt))

	for tableName := range tt {
		tables = append(tables, tableName)
	}

	// @TOOD: return stats (name, size, etc... returned by TableInterface)
	tablesJson, err := json.Marshal(tables)

	if err != nil {
		return err
	}

	return ioutils.WriteDynamicValue[uint32](conn, tablesJson)
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

func WriteTableCfgIO(writer io.Writer, cfg *lsm.CfgTable, cfgType uint8) error {
	var cfgContent []byte
	var err error

	switch cfgType {
	case 0: // toml
		cfgContent, err = toml.Marshal(cfg)

		if err != nil {
			return fmt.Errorf("toml write error: %s", err)
		}
	case 1: // json
		cfgContent, err = json.Marshal(cfg)

		if err != nil {
			return fmt.Errorf("json write error: %s", err)
		}
	default:
		return fmt.Errorf("invalid cfg format: %d", cfgType)
	}

	err = ioutils.WriteDynamic[uint32](writer, cfgContent)

	if err != nil {
		return err
	}

	return nil
}
