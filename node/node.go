package node

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/oboforty/dobo/lsm"
	"github.com/oboforty/dobo/node/socket"

	"github.com/pelletier/go-toml/v2"
)

type Node struct {
	DbPath string
	sock   *socket.TcpSocket

	Tables map[string]lsm.LSMTreeTableInterface
}

type CfgNode struct {
	DbPath string        `toml:"dbpath"`
	Tcp    socket.CfgTcp `toml:"socket"`
}

func NewFromDisc(path string) (*Node, error) {
	cfg, err := ReadNodeCfgFile(path)
	if err != nil {
		log.Fatalf("[Cfg] parse error: %s", err)
		return nil, err
	}

	// apply defaults
	if cfg.DbPath == "" {
		cfg.DbPath = path
	}
	cfg.Tcp.Defaults()

	node := &Node{
		DbPath: cfg.DbPath,
		Tables: make(map[string]lsm.LSMTreeTableInterface),
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

	for _, file := range tableFiles {
		if !file.IsDir() {
			continue
		}

		node.LoadTableFromDisc(file.Name())
	}

	return node, nil
}

// func (n *Node) Table(tableName string) lsm.LSMTreeTableInterface {
// 	v, ok := n.Tables[tableName]

// 	if !ok {
// 		return nil
// 	}

// 	return v
// }

func (node *Node) ListTables() map[string]lsm.LSMTreeTableInterface {
	return node.Tables
}

func (node *Node) Listen() {
	node.sock.Listen()
}

func (node *Node) GetDBPath() string {
	return node.DbPath
}

func (node *Node) LoadTableFromDisc(tableName string) {

	tree, err := lsm.NewFromDisc(filepath.Join(node.DbPath, tableName))

	if err != nil {
		log.Fatalf("[Node] load error: %s", err)
	}

	node.Tables[tree.TableName()] = tree

	// @TODO: print core stats on size & summary
	log.Printf("[Node] loaded table %s", tree.TableName())
}

func ReadNodeCfgFile(path string) (*CfgNode, error) {
	// @TODO: support json, yaml too for file config?

	nodePath := filepath.Join(path, "node.toml")
	cfgContent, err := os.ReadFile(nodePath)

	if err != nil {
		return nil, fmt.Errorf("not found file: %s (%s)", err, nodePath)
	}

	cfg := &CfgNode{}

	err = toml.Unmarshal(cfgContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s (%s)", err, nodePath)
	}

	return cfg, nil
}
