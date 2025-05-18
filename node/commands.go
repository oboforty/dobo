package node

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/lsm/core/ioutils"
	"github.com/oboforty/dobo/node/commands"
)

func (node *Node) handleCommands(conn net.Conn) {
	defer conn.Close()

	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			if err != io.EOF {
				log.Printf("[Cmd] Read error: %s", err)
			}

			// TODO: handle disconnect
			conn.Close()
			break
		}

		var cmd commands.CommandType = b[0]
		cmdDescr, ok := commands.CMD_DESCR[cmd]
		var tableName string
		var tableIF lsm.LSMTreeTableInterface

		if !ok {
			remBytes, err := io.ReadAll(conn)

			if err != nil {
				println("HUHH ???", err)
			}

			log.Printf("[Cmd] Invalid command: %d. Remaining bytes: %v", cmd, remBytes)

			continue
		} else {
			log.Printf("[Cmd] Running command: %s", cmdDescr.Name)
		}

		// Load LSM Table if it's needed
		if cmdDescr.NoTableNeeded {
			err = commands.HandleNodeCommand(node, conn, cmd)

			if err != nil {
				log.Printf("[Cmd] error: %s (%s)", cmdDescr.Name, err)
			}
			continue
		}

		err = ioutils.ReadDynamicValue[uint8](conn, &tableName)
		if err != nil {
			log.Printf("[Cmd] TableName parsing error: %s (%s)", err, cmdDescr.Name)
			continue
		}

		tableIF, ok = node.Tables[tableName]
		if !ok {
			log.Printf("[%s] Table does not exist: %s (%s)", tableName, err, cmdDescr.Name)
			continue
		}

		if commands.GET_ITEM <= cmd && cmd <= commands.DEL_ITEMS {
			// Item commands
			switch table := tableIF.(type) {
			case *lsm.LSMTreeTable[int32]:
				err = commands.HandleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[int64]:
				err = commands.HandleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[float32]:
				err = commands.HandleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[float64]:
				err = commands.HandleItemCommand(table, conn, cmd)
			case *lsm.LSMTreeTable[string]:
				err = commands.HandleItemCommand(table, conn, cmd)
			default:
				err = fmt.Errorf("invalid item command: %d", cmd)
			}
		} else if commands.TABLE_INFO <= cmd && cmd <= commands.DROP_TABLE {
			// Table commands
			err = commands.HandleTableCommand(tableIF, conn, cmd)
		} else {
			err = fmt.Errorf("invalid command: %d", cmd)
		}

		if err != nil {
			log.Printf("[%s] error: %s (%s)", tableName, err, cmdDescr.Name)
			continue
		}
	}
}
