package lsm

import (
	"dobo/lsm/core"
	"dobo/lsm/memtable"
)

type CfgTable struct {
	Name       string
	Items      CfgItems             `toml:"items"`
	Partitions CfgPartitions        `toml:"partitions"`
	MemTable   memtable.CfgMemtable `toml:"memtable"`
	SSTable    CfgSSTable           `toml:"sstable"`
}

//	type LSMTreeMetadata struct {
//		TokenMin int
//		TokenMax int
//	}
type CfgItems struct {
}

type CfgPartitions struct {
	ClusterSize int `toml:"cluster_size"`
}

type CfgSSTable struct {
	BasePath string `toml:"base_path"`

	DynamicValueSerialization core.DynamicValueSerialization `toml:"dynamic_value_serialization"`
}
