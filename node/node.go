package node

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/node/socket"
)

type Node struct {
	DbPath string
	Tables map[string]lsm.LSMTreeTableInterface
	sock   *socket.TcpSocket
}

func NewFromDisc(path string) (*Node, error) {
	cfg, err := ReadNodeConfig(path)
	if err != nil {
		log.Fatalf("[Node] config error: %s", err)
		return nil, err
	}

	// apply defaults
	if cfg.DbPath == "" {
		cfg.DbPath = path
	}
	cfg.Tcp.Defaults()

	node := &Node{
		DbPath: cfg.DbPath,
	}

	node.sock, err = socket.New(&cfg.Tcp, node.handleClient)
	if err != nil {
		log.Fatalf("[Node] socket setup error: %s", err)
		return nil, err
	}

	tableFiles, err := os.ReadDir(cfg.DbPath)
	if err != nil {
		log.Fatalf("[Node] list tables error: %s", err)
		return nil, err
	}

	// node.Tables = make([]lsm.LSMTreeTableInterface, 0, len(tableFiles)-1)

	for _, file := range tableFiles {
		if !file.IsDir() {
			continue
		}

		tree, err := lsm.NewFromDisc(filepath.Join(cfg.DbPath, file.Name()))
		if err != nil {
			log.Fatalf("[Node] load error: %s", err)
			continue
		}

		node.Tables[tree.TableName()] = tree
		// @TODO: print core stats on size & summary
		log.Printf("[Node] loaded table %s", tree.TableName())
	}

	return node, nil
}

func (n *Node) handleClient(conn net.Conn) {
	defer conn.Close()

	// @TODO: Auth or drop

	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			log.Printf("[Server] cmd typ error: %s", err)
			continue
		}

		newCmd, ok := cmds[b[0]]
		if !ok {
			log.Printf("[Server] cmd not found: %b", b[0])
			continue
		}
		cmd := newCmd(conn)

		err = cmd.Run()
		if err != nil {
			log.Fatalf("[%s] unhandled error: %s", getType(cmd), err)
			break
			// or continue?
		}
	}

	// buf := make([]byte, 512)
	// 	n, err = conn.Write(buf[:n])

	log.Printf("[Server]: connection closed %s", conn.RemoteAddr())
}

func (n *Node) Listen() {
	n.sock.Listen()
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
