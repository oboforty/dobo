package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

func SearchKeyInFile2[P comparable](
	filename string,
	startBlockOffset,
	stopBlockOffset int32,
	key P,
) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
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
			return false, err
		}
		keyLength := binary.BigEndian.Uint32(b)

		// @TODO: Temporal
		if keyLength > 65536 {
			println("### ERR 2")

			println("keyLength: ", keyLength)
			println("read bytes:", totalBytes)
			println("MOD 16:", startBlockOffset%16)

			if totalBytes == 0 {
				println("--- BACKTRACKING --- ")

				for i := range 10 {
					println("n-", i)
					file.Seek(int64(startBlockOffset-1), os.SEEK_SET)

					for range 10 {
						_, err = file.Read(b)
						println(fmt.Sprintf("\t %v", b))
					}
					println("------------------------")
				}

				println("------------------------------------------------")
				println("------------------------------------------------")
				println("------------------------------------------------")
			}

			panic("oof")
		}

		// Read Key Bytes (dyn length)
		keyBytes := make([]byte, keyLength)
		_, err = file.Read(keyBytes)
		if err != nil {
			return false, err
		}

		// Read Block offset
		_, err := file.Read(b)
		if err != nil {
			// @TODO: $ITT: debug what happens
			// start reading file backwards
			println("keyBytes: ", fmt.Sprintf("\t %v", keyBytes))
			panic(err)
			return false, err
		}
		blockOffset := binary.BigEndian.Uint32(b)

		_, err = file.Read(b)
		if err != nil {
			return false, err
		}
		interBlockOffset := binary.BigEndian.Uint32(b)

		// @TODO: binary compare instead?
		if bytes.Equal(searchKeyBytes, keyBytes) {
			println("@@@@@@@@@@@@ BINGO @@@@@@@@@@@@")

			println("keylen:", keyLength, "key bytes:", fmt.Sprintf("key bytes:\t %v", keyBytes))
			println("OFFSETS:", blockOffset, interBlockOffset)
			println("read bytes:", totalBytes)
			println("MOD 16:", startBlockOffset%16)

			// var foundKey P
			// err = binary.Read(bytes.NewReader(keyBytes), binary.BigEndian, &foundKey)
			// if err != nil {
			// 	panic(err)
			// } else {
			// 	println("Key:", foundKey)
			// }

			break
		}

		startBlockOffset += int32(3*4 + keyLength)
		totalBytes += 3*4 + keyLength
		if startBlockOffset > stopBlockOffset {
			println("@@@@@@@@ ENDING -------------------", startBlockOffset, stopBlockOffset)
			break
		}
	}

	return true, nil
}
