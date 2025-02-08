package node

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/node/socket"
	"github.com/pelletier/go-toml/v2"
)

type CfgNode struct {
	DbPath string        `toml:"dbpath"`
	Tcp    socket.CfgTcp `toml:"socket"`
}

type Node struct {
	DbPath string
	sock   *socket.TcpSocket

	Tables map[string]lsm.LSMTreeTableInterface
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

	node.sock, err = socket.New(&cfg.Tcp, node.handleCommands)
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

func (n *Node) Table(tableName string) lsm.LSMTreeTableInterface {
	v, ok := n.Tables[tableName]

	if !ok {
		return nil
	}
	return v
}

func (n *Node) Listen() {
	n.sock.Listen()
}

func ReadNodeConfig(path string) (*CfgNode, error) {
	nodePath := filepath.Join(path, "node.toml")
	cfgContent, err := os.ReadFile(nodePath)

	if err != nil {
		log.Fatalf("[Cfg] not found file: %s (%s)", err, nodePath)
		return nil, err
	}

	cfg := &CfgNode{}

	err = toml.Unmarshal(cfgContent, cfg)
	if err != nil {
		log.Fatalf("[Cfg] parse error: %s (%s)", err, nodePath)
		return nil, err
	}

	return cfg, nil
}
