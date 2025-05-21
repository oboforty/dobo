package sstable

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

type SparseIndexHit[P core.PartKeyTypes] interface {
	GetBlockOffset() uint32
	GetInterBlockOffset() uint32
}

type ValueSearchHit struct {
	Value []byte

	// @TODO: add more information?
}

// Fetches the compressed block offset & the offset within the decompressed block for the data file
// using the summary index & 2nd order index files
// func SearchSparseIndex[P core.PartKeyTypes](searchKey P, sparseIndex []SparseIndex[P], filebase string, isFirstOrder bool) (SparseIndexHit[P], error) {
// 	var err error

// 	siHit := BinSearchSparseIndexMemory(searchKey, sparseIndex)

// 	if siHit == nil {
// 		return nil, nil
// 	}

// 	if isFirstOrder {
// 		// the binary search only gave the block offset within the 2nd order (.idx) index file
// 		// Now we search thas file to get the .dat file's block offsets
// 		siHit, err = SearchIndexFile(filebase+".idx", searchKey, siHit.GetBlockOffset())

// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	return siHit, nil
// }

func SearchIndexFile[P core.PartKeyTypes](filename string, searchKey P, startBlockOffset uint32) (SparseIndexHit[P], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var siHit SparseIndexHit[P]

	if startBlockOffset > 0 {
		file.Seek(int64(startBlockOffset), os.SEEK_SET)
	}

	for {
		var startPartKey P
		if err := ioutils.ReadDynamicValue[uint32](file, &startPartKey); err != nil {
			return nil, err
		}

		// Read data block offsets -- only needed for search success
		var blockOffset uint32
		err = binary.Read(file, binary.BigEndian, &blockOffset)
		if err != nil {
			return nil, err
		}
		var interBlockOffset uint32
		err = binary.Read(file, binary.BigEndian, &interBlockOffset)
		if err != nil {
			return nil, err
		}

		if core.UberComparator(startPartKey, searchKey) != 1 {
			// this range is still valid
			siHit = SparseIndex[P]{
				BlockOffset:      blockOffset,
				InterBlockOffset: interBlockOffset,
			}
		} else {
			// startPartKey is larger than the key we're looking for.
			// the previous range is the closest hit for us to begin looking in the block
			break
		}

		// @TODO: stop condition?
	}

	if siHit == nil {
		return nil, fmt.Errorf("index file was scanned till end? :ooo")
	}

	return siHit, nil
}

func SearchDataFileGzipBlock[P core.PartKeyTypes](
	filename string,
	startBlockOffset,
	startInterBlockOffset int32,
	searchKey P,
) (*ValueSearchHit, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if startBlockOffset > 0 {
		file.Seek(int64(startBlockOffset), os.SEEK_SET)
	}

	var blockLength uint32
	err = binary.Read(file, binary.BigEndian, &blockLength)
	if err != nil {
		return nil, err
	}

	// Decompress block
	compReader, err := gzip.NewReader(io.LimitReader(file, int64(blockLength)))
	if err != nil {
		return nil, err
	}
	defer compReader.Close()

	// skip bytes until we get to the relevant offset
	// @TODO: can this be optimzied? e.g. skip with buffer of 1MB?
	// var skipBuffer []byte
	// skipReader := io.LimitReader(compReader, int64(startInterBlockOffset))
	// if _, err = skipReader.Read(skipBuffer); err != nil {
	// 	return nil, err
	// }
	// println("@ OFFSET ", startInterBlockOffset)

	// @TODO: why is this wrong again?
	// now start reading entries until we find our key
	for {
		var partKey P
		if err := ioutils.ReadDynamicValue[uint32](compReader, &partKey); err != nil {
			return nil, err
		}
		// println("@ ", partKey, searchKey)

		value, err := ioutils.ReadDynamic[uint32](compReader)
		if err != nil {
			return nil, err
		}

		if core.UberComparator(partKey, searchKey) == 0 {
			if len(value) == 0 {
				// tombstone found
				value = nil
			}

			return &ValueSearchHit{Value: value}, nil
		}
	}
}

// Linear search first order sparse index
func LinSearchSparseIndexMemory[P core.PartKeyTypes](searchKey P, sparseIndex []SparseIndex[P]) SparseIndexHit[P] {
	var siHit SparseIndexHit[P]

	for _, siRange := range sparseIndex {
		if core.UberComparator(siRange.StartPartKey, searchKey) != 1 {
			// this range is OK to find the searched key,
			siHit = &siRange
		} else {
			// partKey is bigger than this summary's range.
			// the previous start key lies the closest to the key we're looking for
			break
		}
	}

	return siHit
}

// Binary search first order sparse index
func BinSearchSparseIndexMemory[P core.PartKeyTypes](searchKey P, sparseIndex []SparseIndex[P]) SparseIndexHit[P] {
	var siHit SparseIndexHit[P]

	// Binary search first order sparse index
	// @TODO: implement bin search
	low, high := 0, len(sparseIndex)-1
	for low <= high {
		mid := (low + high) / 2
		comp := core.UberComparator(sparseIndex[mid].StartPartKey, searchKey)

		if comp <= 0 {
			// mid is a candidate, but there might be a better (closer) one to the right
			siHit = &sparseIndex[mid]
			low = mid + 1
		} else {
			// current StartPartKey is greater than searchKey, go left
			high = mid - 1
		}
	}

	return siHit
}
