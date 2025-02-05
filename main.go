package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oboforty/dobo/node"
)

func main() {
	var cfgFilePath string
	var err error

	if len(os.Args) <= 1 {
		cfgFilePath, err = os.Getwd()
		if err != nil {
			log.Fatal("[Cfg] unable to get CWD")
		}
	} else {
		cfgFilePath = os.Args[1]
	}

	finfo, err := os.Stat(cfgFilePath)
	if err != nil {
		log.Fatalf("[Cfg] stat error %s", err)
	}

	if !finfo.IsDir() {
		cfgFilePath = filepath.Dir(cfgFilePath)
	}

	// if filepath.dir
	// _, err = config.ReadNodeConfig(cfgFilePath)
	// if err != nil {
	// 	log.Fatalf("[Cfg] Parse error: %s", err)
	// }
	// @TODO: get dbpath from node config?
	dbPath := cfgFilePath

	node, err := node.NewFromDisc(dbPath)

	if err != nil {
		log.Fatalf("[Node] setup error: %s", err)
	}

	node.Listen()
}
