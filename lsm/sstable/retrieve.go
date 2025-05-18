package sstable

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

type IdxSearchHit struct {
	// these are offsets for the dat file!
	BlockOffset      uint32
	InterBlockOffset uint32

	// @TODO: add more information?
}

type ValueSearchHit struct {
	Value []byte

	// @TODO: add more information?
}

// LINEAR SEARCH
// @TODO: implement binary search & use io.LimitReader(keyLength) instead
func SearchOffsetInIndexFile[P core.PartKeyTypes](
	filename string,
	startBlockOffset,
	stopBlockOffset uint32,
	key P,
) (*IdxSearchHit, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if startBlockOffset > 0 {
		file.Seek(int64(startBlockOffset), os.SEEK_SET)
	}

	// part. key to search for
	searchKeyBytes := make([]byte, 0)
	searchKeyBytes, _ = binary.Append(searchKeyBytes, binary.BigEndian, key)

	totalBytes := uint32(0)

	for {
		keyBytes, err := ioutils.ReadDynamic[uint32](file)
		if err != nil {
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

		// println("CMP keys\t", fmt.Sprintf("%v", keyBytes), "\t", fmt.Sprintf("%v", searchKeyBytes))
		if bytes.Equal(searchKeyBytes, keyBytes) {
			// println("keylen:", len(keyBytes), "key bytes:", fmt.Sprintf("key bytes:\t %v", keyBytes))
			// println("OFFSETS:", blockOffset, interBlockOffset)
			// println("read bytes:", totalBytes)
			// println("MOD 16:", startBlockOffset%16)

			return &IdxSearchHit{
				BlockOffset:      blockOffset,
				InterBlockOffset: interBlockOffset,
			}, nil
		}

		keyLength := uint32(len(keyBytes))
		startBlockOffset += 3*4 + keyLength
		totalBytes += 3*4 + keyLength
		if startBlockOffset > stopBlockOffset {
			break
		}
	}

	return nil, nil
}

// @TODO: implement binary search
func SearchDataFileGzipBlock[P core.PartKeyTypes](
	filename string,
	startBlockOffset,
	startInterBlockOffset int32,
	key P,
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

	decompressed, err := io.ReadAll(compReader)
	if err != nil {
		return nil, err
	}

	// @TODO: somewhere here handle Tombstone entries?

	valueLengthBytes := decompressed[startInterBlockOffset : startInterBlockOffset+4]
	valueLength := binary.BigEndian.Uint32(valueLengthBytes)
	if valueLength > ioutils.MaxUIntDataLength {
		return nil, fmt.Errorf("invalid value length found")
	}

	if valueLength == 0 {
		// tombstone entry, deleted
		return &ValueSearchHit{
			Value: nil,
		}, nil
	}

	value := decompressed[startInterBlockOffset+4 : startInterBlockOffset+4+int32(valueLength)]

	return &ValueSearchHit{
		Value: value,
	}, nil
}
