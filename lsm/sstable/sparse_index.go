package sstable

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

// Can represent a record in the index or an index summary file (.sum)
// Represents a range in a sparse index file. This can be:
// 1) summary file (sst_name.sum), or first order index, also cached in memory
// 2) index file (sst_name.idx), or second order index
type SparseIndex[P core.PartKeyTypes] struct {
	Id uint32
	// The first partition key in this range
	StartPartKey P
	// the byte offset within the file for this range.
	// for data file this marks the start of the compressed block
	BlockOffset uint32
	// only relevant when 2nd order index file (.idx) is loaded into memory.
	// this points to the offset within the decompressed block
	InterBlockOffset uint32
}

func (s SparseIndex[P]) GetBlockOffset() uint32 {
	return s.BlockOffset
}

func (s SparseIndex[P]) GetInterBlockOffset() uint32 {
	return s.InterBlockOffset
}

func ReadSummaryFile[P core.PartKeyTypes](ss *SSTable[P]) error {
	file, err := os.OpenFile(ss.FileBase()+".sum", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 0th part - default statistics
	ss.Statistics = map[string]int{
		// "nr": 0, // number of records
	}

	// 1st part - (ascii) metadata
	ss.Metadata = make(map[string]string, 0)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "------" {
			break
		}

		spl := strings.Split(line, "=")
		ss.Metadata[spl[0]] = spl[1]
	}

	// 2nd part - table statistics
	for scanner.Scan() {
		line := scanner.Text()

		if line == "------" {
			break
		}

		spl := strings.Split(line, "=")
		sv, err := strconv.ParseInt(spl[1], 10, 64)

		if err != nil {
			if err == io.EOF {
				break
			}

			continue
		}
		ss.Statistics[spl[0]] = int(sv)
	}

	// 3rd part - (binary) key ranges for .idx
	ss.sparseIndex = make([]SparseIndex[P], 0)

	if ss.Metadata["sparse_type"] == "sum" {
		// summary contains a sparse index over the .idx file to speed up file IO
		for {
			sum := SparseIndex[P]{}

			err = ioutils.ReadDynamicValue[uint32](file, &sum.StartPartKey)
			if err != nil {
				if err == io.EOF {
					// EOF really should only occur here
					break
				}
				return err
			}

			err = binary.Read(file, binary.BigEndian, sum.BlockOffset)
			if err != nil {
				return err
			}

			ss.sparseIndex = append(ss.sparseIndex, sum)
		}
	} else {
		// load index file into memory, as it's small enough.
		// binary search over idx will reduce file IO further

		idx_file, err := os.OpenFile(ss.FileBase()+".idx", os.O_RDONLY, 0644)
		if err != nil {
			return err
		}
		defer idx_file.Close()

		for {
			sum := SparseIndex[P]{}

			err = ioutils.ReadDynamicValue[uint32](idx_file, &sum.StartPartKey)
			if err != nil {
				if err == io.EOF {
					// EOF really should only occur here
					break
				}
				return err
			}

			err = binary.Read(file, binary.BigEndian, sum.BlockOffset)
			if err != nil {
				return err
			}

			err = binary.Read(file, binary.BigEndian, sum.InterBlockOffset)
			if err != nil {
				return err
			}

			ss.sparseIndex = append(ss.sparseIndex, sum)
		}
	}

	return nil
}

func WriteSummaryFile[P core.PartKeyTypes](filename string, ss *SSTable[P], includeIdx bool) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// 1st part - metadata
	for k, v := range ss.Metadata {
		_, err := fmt.Fprintf(file, "%s=%s\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(file, "------")
	if err != nil {
		return err
	}

	// 2nd part - statistics
	for k, v := range ss.Statistics {
		_, err := fmt.Fprintf(file, "%s=%d\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(file, "------")
	if err != nil {
		return err
	}

	if includeIdx {
		// 3rd part - binary key ranges for .idx
		for _, sum := range ss.sparseIndex {
			err := ioutils.WriteDynamicValue[uint32](file, sum.StartPartKey)
			if err != nil {
				return err
			}

			err = binary.Write(file, binary.BigEndian, sum.BlockOffset)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
