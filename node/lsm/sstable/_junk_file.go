package sstable

import (
	"fmt"
	"io"
	"os"
)

func readBlock(filePath string, start, end int64) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Seek to the start position
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek in file: %v", err)
	}

	size := end - start
	if size < 0 {
		return nil, fmt.Errorf("invalid range: start (%d) is greater than end (%d)", start, end)
	}

	// Read the block
	buffer := make([]byte, size)
	_, err = io.ReadFull(file, buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read block: %v", err)
	}

	return buffer, nil
}

func createEmptyFile(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		// fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

}
