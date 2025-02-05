package node

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/oboforty/dobo/node/socket"
)

type CfgNode struct {
	DbPath string        `toml:"dbpath"`
	Tcp    socket.CfgTcp `toml:"socket"`
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
