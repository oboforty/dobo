package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oboforty/dobo/lsm"
)

func main() {
	var cfgFilePath string
	var err error

	if len(os.Args) <= 1 {
		cfgFilePath, err = os.Getwd()
		if err != nil {
			log.Fatal("[Cfg] Unable to get CWD")
		}
	} else {
		cfgFilePath = os.Args[1]
	}

	finfo, _ := os.Stat(cfgFilePath)
	// if err != nil {
	// log.Fatal("[Cfg] Unable to get CWD")

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

	tables, err := os.ReadDir(dbPath)
	if err != nil {
		log.Fatalf("[Cfg] Parse error: %s", err)
	}

	for _, file := range tables {
		if !file.IsDir() {
			continue
		}

		tree, err := lsm.NewFromDisc(filepath.Join(dbPath, file.Name()))
		if err != nil {
			log.Fatalf("[LSM] Load error: %s", err)
		}

		// @TODO: print core stats on size & summary
		log.Printf("[LSM] Loaded table %s", tree.TableName())
	}
}
