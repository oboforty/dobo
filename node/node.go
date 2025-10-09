package node

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
		return nil, fmt.Errorf("parse error: %s", err)
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
		return nil, fmt.Errorf("socket error: %s", err)
	}

	tableFiles, err := os.ReadDir(cfg.DbPath)
	if err != nil {
		return nil, fmt.Errorf("list tables dir error: %s", err)
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

func (node *Node) RunServer() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// fmt.Println("Running. Press Ctrl+C to exit...")

	node.sock.Listen()

	<-ctx.Done()
	node.CloseServer()
}

func (node *Node) CloseServer() {
	tt := node.ListTables()

	// has_err := false

	for _, table := range tt {
		err := table.FlushMemToDisc()

		if err != nil {
			//("[%s] flush error: %s", table.TableName(), err)

		}
	}
}

func (node *Node) GetDBPath() string {
	return node.DbPath
}

func (node *Node) LoadTableFromDisc(tableName string) {

	tree, err := lsm.NewFromDisc(filepath.Join(node.DbPath, tableName))

	if err != nil {
		slog.Error(fmt.Sprintf("[%s] load error: %s", tableName, err))
	}

	node.Tables[tree.TableName()] = tree

	// @TODO: print core stats on size & summary
	slog.Info(fmt.Sprintf("[%s] loaded from disc", tableName))
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
