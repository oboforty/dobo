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
	// var skipBuffer []byte
	// skipReader := io.LimitReader(compReader, int64(startInterBlockOffset))
	// if _, err = skipReader.Read(skipBuffer); err != nil {
	// 	return nil, err
	// }

	// Skip to the correct position within the block
	if startInterBlockOffset > 0 {
		skipBuffer := make([]byte, 1024) // 1KB buffer for skipping
		remaining := int64(startInterBlockOffset)
		for remaining > 0 {
			toRead := int64(len(skipBuffer))
			if toRead > remaining {
				toRead = remaining
			}
			n, err := compReader.Read(skipBuffer[:toRead])
			if err != nil && err != io.EOF {
				return nil, err
			}
			remaining -= int64(n)
			if n == 0 {
				break
			}
		}
	}

	// Read entries until we find our key or reach EOF
	for {
		var partKey P
		if err := ioutils.ReadDynamicValue[uint32](compReader, &partKey); err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("key not found in block")
			}
			return nil, err
		}

		value, err := ioutils.ReadDynamic[uint32](compReader)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("value not found in block")
			}
			return nil, err
		}

		// println("@ RE>> ", partKey)
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

// DataFileGzipBlockIterator represents an iterator over a gzipped data file containing blocks
type DataFileGzipBlockIterator[P core.PartKeyTypes] struct {
	file        *os.File
	compReader  *gzip.Reader
	blockLength uint32
	done        bool
}

func NewDataFileGzipBlockIterator[P core.PartKeyTypes](filename string) (*DataFileGzipBlockIterator[P], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &DataFileGzipBlockIterator[P]{
		file: file,
	}, nil
}

// Next reads the next key-value pair from the current block
// Returns false when there are no more entries to read
func (it *DataFileGzipBlockIterator[P]) Next() (*core.Item[P], error) {
	if it.done {
		return nil, io.EOF
	}

	// read the next gzip block
	if it.compReader == nil {
		if err := it.readNextBlock(); err != nil {
			if err == io.EOF {
				it.done = true
				return nil, err
			}
			return nil, err
		}
	}

	var partKey P
	if err := ioutils.ReadDynamicValue[uint32](it.compReader, &partKey); err != nil {
		if err == io.EOF {
			// End of current block, try next block
			it.compReader.Close()
			it.compReader = nil
			return it.Next()
		}
		return nil, err
	}

	value, err := ioutils.ReadDynamic[uint32](it.compReader)
	if err != nil {
		println("@@ baj van more", err.Error(), partKey)
		return nil, err
	}

	return &core.Item[P]{
		PartKey: partKey,
		Value:   value,
	}, nil
}

// readNextBlock reads the next compressed block from the file
func (it *DataFileGzipBlockIterator[P]) readNextBlock() error {
	// block length
	err := binary.Read(it.file, binary.BigEndian, &it.blockLength)
	if err != nil {
		return err
	}

	it.compReader, err = gzip.NewReader(io.LimitReader(it.file, int64(it.blockLength)))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return nil
}

func (it *DataFileGzipBlockIterator[P]) Close() error {
	if it.compReader != nil {
		it.compReader.Close()
	}
	return it.file.Close()
}
