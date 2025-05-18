package ioutils

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type BlockReader interface {
	io.Reader
	Close() error
}

func NewBlockReader(filename string, blockSize uint32, compressMethod bool) (BlockReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if compressMethod {
		// @TODO: also add buffered IO for this!
		return &compressedBlockReader{
			blockSize: int(blockSize),
			// filename:  filename,
			file:               file,
			decompressedBuffer: bytes.Buffer{},
		}, nil
	} else {
		return nil, fmt.Errorf("!! non-compressed files not supported yet")
	}
}

type compressedBlockReader struct {
	file               *os.File
	decompressedBuffer bytes.Buffer
	currentBlock       uint32
	blockSize          int
}

func (br *compressedBlockReader) Read(p []byte) (int, error) {
	if br.decompressedBuffer.Len() == 0 {
		err := br.loadNextBlock()
		if err != nil {
			return 0, err
		}
	}

	return br.decompressedBuffer.Read(p)
}

func (br *compressedBlockReader) loadNextBlock() error {
	var blockSize uint32
	err := binary.Read(br.file, binary.BigEndian, &blockSize)
	if err != nil {
		return err
	}

	compressedBlock := make([]byte, blockSize)
	_, err = io.ReadFull(br.file, compressedBlock)
	if err != nil {
		return err
	}

	decompressedData, err := decompressBlock(compressedBlock)
	if err != nil {
		return err
	}

	br.decompressedBuffer.Write(decompressedData)
	br.currentBlock += blockSize + 4

	return nil
}

func decompressBlock(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reader.Close()
	return decompressedData, nil
}

func (br *compressedBlockReader) Close() error {
	return br.file.Close()
}
