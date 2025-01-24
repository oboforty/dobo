package sstable

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const MaxUint = ^uint32(0) - 1

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

// Reads dataLength (uint32) and data of size `dataLength`
type DynamicLengthReader struct {
	reader  io.Reader
	padding uint32 // offset to ignore after value

}

// LINEAR SEARCH
// @TODO: implement binary & use io.LimitReader(keyLength) instead
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

	if startBlockOffset > 0 {
		file.Seek(int64(startBlockOffset), os.SEEK_SET)
	}

	// part. key to search for
	searchKeyBytes := make([]byte, 0)
	searchKeyBytes, _ = binary.Append(searchKeyBytes, binary.BigEndian, key)

	totalBytes := uint32(0)

	b := make([]byte, 4)

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

func SearchDataFileGzipBlock[P comparable](
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

	b := make([]byte, 4)
	// Read block length
	_, err = file.Read(b)
	if err != nil {
		return nil, err
	}
	blockLength := binary.BigEndian.Uint32(b)
	if blockLength > MaxUint {
		return nil, fmt.Errorf("invalid block Length found")
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

	valueLengthBytes := decompressed[startInterBlockOffset : startInterBlockOffset+4]
	valueLength := binary.BigEndian.Uint32(valueLengthBytes)
	if valueLength > MaxUint {
		return nil, fmt.Errorf("invalid value length found")
	}

	value := decompressed[startInterBlockOffset+4 : startInterBlockOffset+4+int32(valueLength)]

	return &ValueSearchHit{
		Value: value,
	}, nil
}
