package lsm

import (
	"cmp"
	"dobo/lsm/core"
	"dobo/lsm/memtable"
	"iter"
)

type Table[P cmp.Ordered, V any] interface {
	Get(P) *core.ItemQuery[P, V]
	Upsert(*core.ItemWrite[P, V])
	// Delete(interface{})
}

type MemTable[P cmp.Ordered, V any] interface {
	Table[P, V]

	ByteSize() uint
	// Size() uint
	IsFull() bool
	Clear()

	ItemIterator() iter.Seq[*core.ItemQuery[P, V]]
}

type SSTable[P cmp.Ordered, V any] interface {
	Table[P, V]
}

type WAL interface {
}

// GET, QUERY - bulk query, can filter,
// PUT/UPSERT, UPDATE - partial update, can be bulk
// REMOVE - soft delete, DELETE - hard delete,
// CREATE - fails if exists

// CQRS: (GET, QUERY), (CREATE, UPSERT, UPDATE, REMOVE, DELETE), (BALANCE-INDEX, SETCONFIG, CREATE-TABLE, DELETE-TABLE, REPLICATE-TABLE, CREATE-PARTITION, DELETE-PARTITION, REPARTITION-TABLE)

type LSMTreeTable[P cmp.Ordered, V any] struct {
	Siblings []*LSMTreeTable[P, V]

	// Components
	Memtable MemTable[P, V]
	Sstables []SSTable[P, V] // latest table is the newest
	Wal      WAL
}

func New[P cmp.Ordered, V any](cfg *CfgTable) (*LSMTreeTable[P, V], error) {
	t := &LSMTreeTable[P, V]{}

	switch cfg.MemTable.Type {
	case MEMTYPE_REDBLACK:
	default:
		memt, err := memtable.NewRedBlack[P, V]()

		if err != nil {
			return nil, err
		}

		t.Memtable = memt
	}

	return t, nil
}

func (t *LSMTreeTable[P, V]) Get(partKey P) *core.ItemQuery[P, V] {
	var item *core.ItemQuery[P, V]

	item = t.Memtable.Get(partKey)
	if item != nil {
		return item
	}

	for i := len(t.Sstables) - 1; i >= 0; i-- {
		item = t.Sstables[i].Get(partKey)

		// @TODO: handle overflow & uint64?
		if item != nil {
			item.FoundSSIdx = uint32(i)
			return item
		}
	}

	return item
}

func (t *LSMTreeTable[P, V]) Upsert(item *core.ItemWrite[P, V]) {
	t.Memtable.Upsert(item)

	if t.Memtable.IsFull() {
		// Mem Tree has grown to its limit. Trigger an SSTable write task

		// @TODO: Log
		println("Flushing MemTable")

		var memt = t.Memtable

		go FlushMemTable(memt)
	}

}

// func (d *LSMTreeTable) Delete(partKey P) {
// 	d.Memtable.Delete(partKey)
// }
