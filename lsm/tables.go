package lsm

import (
	"log"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

type LSMTreeTable[P core.PartKeyTypes] struct {
	cfg             CfgTable
	partKeyTypeInfo core.TypeInfo

	// Components
	MemTable MemTable[P]
	SSTables []SSTable[P] // latest table is the newest
	Wal      WAL
}

type SSTable[P core.PartKeyTypes] interface {
	Get(P) *core.ItemQuery[P]

	GetGenerationId() int
}

type MemTable[P core.PartKeyTypes] interface {
	sstable.IterableTable[P]

	Get(P) *core.ItemQuery[P]
	Upsert(*core.Item[P])
	Delete(P) bool

	ByteSize() uint64
	IsFull() bool
	Clear()
}

type WAL interface {
}

func New[P core.PartKeyTypes](cfg *CfgTable) (*LSMTreeTable[P], error) {
	cfg.ApplyDefaults()

	t := &LSMTreeTable[P]{
		cfg: *cfg,
	}

	typeInfo, err := core.GetTypeInfo[P]()
	if err != nil {
		return nil, err
	}
	t.partKeyTypeInfo = *typeInfo

	t.createMemtable()
	t.loadSSTables()

	return t, nil
}

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

func (t *LSMTreeTable[P]) Upsert(item *core.Item[P]) bool {
	t.MemTable.Upsert(item)

	// Mem Tree has grown to its limit. recommend client to trigger flush task
	return t.MemTable.IsFull()
}

// @TODO: Tombstone
func (t *LSMTreeTable[P]) Delete(partKey P) {
	t.MemTable.Delete(partKey)
}

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
	log.Println("[SST] Flushing mem to disc, size: ", t.MemTable.ByteSize())

	memtOld := t.MemTable
	t.createMemtable()
	defer memtOld.Clear()

	// @TODO: pass db path?

	// @TODO: new & pass cfg in one
	ss := sstable.New[P](
		&t.cfg.SSTable,
		t.cfg.Name,
		t.partKeyTypeInfo,
		t.CurrentGenerationId()+1,
	)
	t.SSTables = append(t.SSTables, ss)

	ss.Statistics["flushed_at_mem_size"] = int(memtOld.ByteSize())

	return ss.WriteToDisc(memtOld)
}

func (t *LSMTreeTable[P]) TableName() string {
	return t.cfg.Name
}
