package sstable

import (
	"iter"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
)

type COMPACTION_STRATEGY uint8

const (
	COMPACT_SAME_SIZE COMPACTION_STRATEGY = iota
	COMPACT_SAME_LVL
	COMPACT_TIME_WINDOW
)

// Yields a new table by merging the data of two tables.
// This iterator can be fed into sstable.WriteToDisc
type CompactionIterable[P core.PartKeyTypes] struct {
	table1 *SSTable[P]
	table2 *SSTable[P]
}

// Reads both files in a merge sort manner
// Removes expired tombstones & duplicates
func (c *CompactionIterable[P]) ItemIterator() iter.Seq[*core.ItemQuery[P]] {

	// @TODO: ITT....

	return func(yield func(*core.ItemQuery[P]) bool) {
		item := &core.ItemQuery[P]{}

		if !yield(item) {
			return
		}
	}
}

func (c *CompactionIterable[P]) Len() uint32 {
	return uint32(c.table1.Statistics["records"] + c.table2.Statistics["records"])
}

func (c *CompactionIterable[P]) TotalKeySize() uint64 {
	return uint64(c.table1.Statistics["total_key_size"]) + uint64(c.table2.Statistics["total_key_size"])
}
func (c *CompactionIterable[P]) TotalValueSize() uint64 {
	return uint64(c.table1.Statistics["total_data_size"]) + uint64(c.table2.Statistics["total_data_size"])
}

func CompactTables[P core.PartKeyTypes](table1 *SSTable[P], table2 *SSTable[P]) (*SSTable[P], error) {
	// file, err := os.Open(table1.FileBase())
	// settings are ought to be the same
	var table SSTable[P] = *table1
	table.Statistics = map[string]int{}
	table.Metadata = core.CloneMap(table1.Metadata)

	// @TODO: edit meta/stats -- generation ID
	table.bloom = bloom.New(bloom.CfgBloomFilter{
		FalsePositiveRate: table1.bloom.FalsePositiveRate(),
	})

	merge := &CompactionIterable[P]{
		table1: table1,
		table2: table2,
	}
	if err := table.WriteToDisc(merge); err != nil {
		return nil, err
	}

	// only matters for recently flushed SSTables
	// delete(table.Statistics, "flushed_at_mem_size")

	return &table, nil
}
