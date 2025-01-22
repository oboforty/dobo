package lsm

import (
	"dobo/lsm/memtable"
	"dobo/lsm/sstable"
)

type CfgTable struct {
	Name       string
	Partitions CfgPartitions        `toml:"partitions"`
	MemTable   memtable.CfgMemtable `toml:"memtable"`
	SSTable    sstable.CfgSSTable   `toml:"sstable"`
}

type CfgPartitions struct {
	ClusterSize int `toml:"cluster_size"`
	// TokenMin int
	// TokenMax int
}
