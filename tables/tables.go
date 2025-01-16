package tables

import "dobo/tables/core"

type Table interface {
	Get(interface{}) *core.Item
	Upsert(interface{}, interface{})
	Delete(interface{})
}

type MemTable interface {
	Table

	Init(core.DataType)
}

type SSTable interface {
	Table

	Init(core.DataType)
}

type WAL interface {
}

type LSMTable struct {
	Metadata *core.TableMetadata

	Siblings []*LSMTable

	// Components
	Memtable MemTable
	Sstables []SSTable
	Wal      WAL
}

func (d *LSMTable) Init(keyType core.DataType) {
	d.Memtable.Init(keyType)
}

func (d *LSMTable) Get(partKey interface{}) *core.Item {
	var item *core.Item

	item = d.Memtable.Get(partKey)
	if item != nil {
		return item
	}

	for i, sstable := range d.Sstables {
		item = sstable.Get(partKey)

		// @TODO: handle overflow & uint64?
		if item != nil {
			item.FoundSSIdx = uint32(i)
			return item
		}
	}

	return item
}

// func (d *LSMTable) Upsert(partKey interface{}, value interface{}) {
// 	d.tree.Put(partKey, value)
// }

// func (d *LSMTable) Delete(partKey interface{}) {
// 	d.Memtable.Delete(partKey)
// }
