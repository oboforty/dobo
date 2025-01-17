package lsm

type CfgTable struct {
	Name       string
	Items      CfgItems      `toml:"items"`
	Partitions CfgPartitions `toml:"partitions"`
	MemTable   CfgMemtable   `toml:"memtable"`
	SSTable    CfgSSTable    `toml:"sstable"`
}

//	type LSMTreeMetadata struct {
//		TokenMin int
//		TokenMax int
//	}
type CfgItems struct {
	// PartKeyType core.DataType
	// SortKeyType core.DataType
}

type CfgPartitions struct {
	ClusterSize int `toml:"cluster_size"`
}

type CfgMemtableType = string

const (
	MEMTYPE_REDBLACK CfgMemtableType = "redblack"
	MEMTYPE_AVL      CfgMemtableType = "avl"
	MEMTYPE_SKIPLIST CfgMemtableType = "skiplist"
)

type CfgMemtable struct {
	Type CfgMemtableType `toml:"type"`
}

type CfgSSTable struct {
}
