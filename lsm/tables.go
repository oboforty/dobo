package lsm

import (
	"cmp"
	"iter"
	"log"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

type MemTable[P cmp.Ordered] interface {
	Get(P) *core.ItemQuery[P]
	Upsert(*core.ItemWrite[P])
	// Delete(interface{})

	ByteSize() uint
	IsFull() bool
	Clear()

	ItemIterator() iter.Seq[*core.ItemQuery[P]]
}

type SSTable[P cmp.Ordered] interface {
	Get(P) *core.ItemQuery[P]

	GetGenerationId() int
}

type WAL interface {
}

type CfgTable struct {
	Name string
	// Partitions CfgPartitions        `toml:"partitions"`
	MemTable memtable.CfgMemtable `toml:"memtable"`
	SSTable  sstable.CfgSSTable   `toml:"sstable"`
}

// type CfgPartitions struct {
// 	ClusterSize int `toml:"cluster_size"`
// 	// TokenMin int
// 	// TokenMax int
// }

type LSMTreeTable[P cmp.Ordered] struct {
	TableName       string
	partKeyTypeInfo core.TypeInfo
	cfg             CfgTable

	// Components
	MemTable MemTable[P]
	SSTables []SSTable[P] // latest table is the newest
	Wal      WAL
}

func New[P cmp.Ordered](cfg *CfgTable) (*LSMTreeTable[P], error) {
	t := &LSMTreeTable[P]{
		cfg: *cfg,
	}

	typeInfo, err := core.GetTypeInfo[P]()
	if err != nil {
		return nil, err
	}
	t.partKeyTypeInfo = *typeInfo

	if cfg.MemTable.Type == "" {
		cfg.MemTable.Type = memtable.MEMTYPE_REDBLACK
	}

	t.createMemtable()
	t.loadSSTables()

	return t, nil
}

// GET, QUERY - bulk query, can filter,
// PUT/UPSERT, UPDATE - partial update, can be bulk
// REMOVE - soft delete, DELETE - hard delete,
// CREATE - fails if exists

// CQRS: (GET, QUERY), (CREATE, UPSERT, UPDATE, REMOVE, DELETE), (BALANCE-INDEX, SETCONFIG, CREATE-TABLE, DELETE-TABLE, REPLICATE-TABLE, CREATE-PARTITION, DELETE-PARTITION, REPARTITION-TABLE)

func (t *LSMTreeTable[P]) Get(partKey P) *core.ItemQuery[P] {
	var item *core.ItemQuery[P]

	item = t.MemTable.Get(partKey)
	if item != nil {
		return item
	}

	for i := len(t.SSTables) - 1; i >= 0; i-- {
		item = t.SSTables[i].Get(partKey)

		// @TODO: handle overflow & uint64?
		if item != nil {
			item.FoundSSIdx = uint32(i)
			return item
		}
	}

	return item
}

func (t *LSMTreeTable[P]) Upsert(item *core.ItemWrite[P]) bool {
	t.MemTable.Upsert(item)

	// Mem Tree has grown to its limit. recommend client to trigger flush task
	return t.MemTable.IsFull()
}

// @TODO: Tombstone
// func (d *LSMTreeTable) Delete(partKey P) {
// 	d.Memtable.Delete(partKey)
// }

func (t *LSMTreeTable[P]) CurrentGenerationId() int {
	l := len(t.SSTables)

	if l == 0 {
		return 0
	}

	return t.SSTables[l-1].GetGenerationId()
}

func (t *LSMTreeTable[P]) createMemtable() {
	switch t.cfg.MemTable.Type {
	case memtable.MEMTYPE_REDBLACK:
		t.MemTable = memtable.NewRedBlack[P](
			&t.cfg.MemTable,
			&t.partKeyTypeInfo,
		)
	}
}

func (t *LSMTreeTable[P]) FlushMemToDisc() error {
	// @TODO: Log

	log.Println("[SST] Flushing MemTable, size: ", t.MemTable.ByteSize())

	memtOld := t.MemTable
	t.createMemtable()
	defer memtOld.Clear()

	// @TODO: new & pass cfg in one
	ss := sstable.New[P](
		&t.cfg.SSTable,
		t.TableName,
		t.partKeyTypeInfo,
		t.CurrentGenerationId()+1,
	)
	t.SSTables = append(t.SSTables, ss)

	return ss.WriteToDisc(memtOld)
}
