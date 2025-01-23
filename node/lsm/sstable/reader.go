package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const MaxUint = ^uint32(0) - 1

type IdxSearchHit struct {
	// these are offsets for the dat file!
	BlockOffset      uint32
	InterBlockOffset uint32

	// @TODO: add more information?
}

// LINEAR SEARCH
// @TODO: implement binary
func SearchOffsetInIndexFile[P comparable](
	filename string,
	startBlockOffset,
	stopBlockOffset int32,
	key P,
) (*IdxSearchHit, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	b := make([]byte, 4)

	if startBlockOffset > 0 {
		file.Seek(int64(startBlockOffset), os.SEEK_SET)
	}

	// part. key to search for
	searchKeyBytes := make([]byte, 0)
	searchKeyBytes, _ = binary.Append(searchKeyBytes, binary.BigEndian, key)

	totalBytes := uint32(0)

	for {
		// Read Key Length int32
		_, err = file.Read(b)
		if err != nil {
			return nil, err
		}
		keyLength := binary.BigEndian.Uint32(b)

		if keyLength > MaxUint {
			return nil, fmt.Errorf("Invalid keylength found. KeyLength= %d, Block offset % 16 = %d, Total read bytes = %d", keyLength, startBlockOffset, totalBytes)
		}

		// Read Key Bytes (dyn length)
		keyBytes := make([]byte, keyLength)
		_, err = file.Read(keyBytes)
		if err != nil {
			return nil, err
		}

		// Read data block offsets -- only needed for search success
		_, err := file.Read(b)
		if err != nil {
			return nil, err
		}
		blockOffset := binary.BigEndian.Uint32(b)
		_, err = file.Read(b)
		if err != nil {
			return nil, err
		}
		interBlockOffset := binary.BigEndian.Uint32(b)

		if bytes.Equal(searchKeyBytes, keyBytes) {
			// println("keylen:", keyLength, "key bytes:", fmt.Sprintf("key bytes:\t %v", keyBytes))
			// println("OFFSETS:", blockOffset, interBlockOffset)
			// println("read bytes:", totalBytes)
			// println("MOD 16:", startBlockOffset%16)

			return &IdxSearchHit{
				BlockOffset:      blockOffset,
				InterBlockOffset: interBlockOffset,
			}, nil
		}

		startBlockOffset += int32(3*4 + keyLength)
		totalBytes += 3*4 + keyLength
		if startBlockOffset > stopBlockOffset {
			break
		}
	}

	return nil, nil
}
