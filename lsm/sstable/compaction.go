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

	file1 *DataFileGzipBlockIterator[P]
	file2 *DataFileGzipBlockIterator[P]
}

// Reads both files in a merge sort manner
// Removes expired tombstones & duplicates
func (c *CompactionIterable[P]) ItemIterator() iter.Seq[*core.Item[P]] {
	return func(yield func(*core.Item[P]) bool) {
		item1, err1 := c.file1.Next()
		item2, err2 := c.file2.Next()

		for err1 == nil || err2 == nil {
			var val1, val2 P

			if item1 != nil {
				val1 = item1.PartKey
			}
			if item2 != nil {
				val2 = item2.PartKey
			}
			print("@ ", val1, " ", val2)
			if err1 != nil {
				// Only f2 has remaining entries
				print(" -> ", item2.PartKey, "\n")
				if !yield(item2) {
					return
				}
				item2, err2 = c.file2.Next()
			} else if err2 != nil {
				// Only f1 has remaining entries
				print(" -> ", item1.PartKey, "\n")
				if !yield(item1) {
					return
				}
				item1, err1 = c.file1.Next()
			} else {
				if core.UberComparator(item1.PartKey, item2.PartKey) != 1 {
					print(" -> ", item1.PartKey, "\n")
					if !yield(item1) {
						return
					}
					item1, err1 = c.file1.Next()
				} else {
					print(" -> ", item2.PartKey, "\n")
					if !yield(item2) {
						return
					}
					item2, err2 = c.file2.Next()
				}
			}
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

func (c *CompactionIterable[P]) Open() error {
	var err error

	if c.file1, err = NewDataFileGzipBlockIterator[P](c.table1.FileBase() + ".dat"); err != nil {
		return nil
	} else if c.file2, err = NewDataFileGzipBlockIterator[P](c.table2.FileBase() + ".dat"); err != nil {
		return nil
	}

	return nil
}

func (c *CompactionIterable[P]) Close() error {
	if err := c.file1.Close(); err != nil {
		return err
	} else if err := c.file2.Close(); err != nil {
		return err
	}

	return nil
}

func CompactTables[P core.PartKeyTypes](table1 *SSTable[P], table2 *SSTable[P], generationId int) (*SSTable[P], error) {
	// settings are ought to be the same
	var table SSTable[P] = *table1
	table.Statistics = map[string]int{}
	table.Metadata = core.CloneMap(table1.Metadata)
	table.GenerationId = generationId

	// @TODO: edit meta/stats -- generation ID
	table.bloom = bloom.New(bloom.CfgBloomFilter{
		FalsePositiveRate: table1.bloom.FalsePositiveRate(),
	})

	merge := &CompactionIterable[P]{
		table1: table1,
		table2: table2,
	}

	if err := merge.Open(); err != nil {
		return nil, err
	}
	defer merge.Close()

	if err := table.WriteToDisc(merge); err != nil {
		return nil, err
	}

	// only matters for recently flushed SSTables
	// delete(table.Statistics, "flushed_at_mem_size")

	return &table, nil
}
