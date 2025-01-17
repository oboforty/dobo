package main

import (
	"dobo/lsm"
	"log"

	"github.com/pelletier/go-toml/v2"
)

var tomlData = `
name = "mydb"

[items]
key_type = "int32"
sort_key_type = "int32"
# "json" | "bytes" | "string"
value_type = "json"

[partitions]
cluster_size = 1

[memtable]
# "redblack" | "avl" | "skiplist"
type = "redblack"

[sstable]
# "sst_io" | "parquet"
type = "sst_io"
# "summary" | "lru_cache" | "none"
lookup_aid = "lru_cache"

# TODO:
# crc
`

func main() {
	var conf lsm.CfgTable
	if err := toml.Unmarshal([]byte(tomlData), &conf); err != nil {
		log.Fatal(err)
	}
	log.Printf("title: %s", conf.Name)
	log.Printf("Feature 1: %#v", conf.Items.PartKeyType)
	log.Printf("Feature 2: %#v", conf.MemTable.Type)
}
