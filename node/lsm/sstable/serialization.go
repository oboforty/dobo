package sstable

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"os"
)

// compressedBlockWriter handles writing compressed blocks of data to a file with buffering
type compressedBlockWriter struct {
	// Maximum size of each block
	blockSize    int
	filename     string
	file         *os.File
	buffer       bytes.Buffer
	currentBlock int32
}

func NewCompressedWriter(filename string, blockSize uint32) (*compressedBlockWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &compressedBlockWriter{
		blockSize: int(blockSize),
		filename:  filename,
		file:      file,
		buffer:    bytes.Buffer{},
	}, nil
}

func (bc *compressedBlockWriter) CompressBlock(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (bc *compressedBlockWriter) Write(data []byte) (int32, error) {
	bc.buffer.Write(data)

	if bc.buffer.Len() >= bc.blockSize {
		return bc.currentBlock, bc.flushBuffer()
	}
	return bc.currentBlock, nil
}

func (bc *compressedBlockWriter) flushBuffer() error {
	if bc.buffer.Len() == 0 {
		return nil
	}

	compressedBlock, err := bc.CompressBlock(bc.buffer.Bytes())
	if err != nil {
		return err
	}

	// Write the size of the compressed block first
	blockSize := int32(len(compressedBlock))
	err = binary.Write(bc.file, binary.LittleEndian, blockSize)
	if err != nil {
		return err
	}

	// Write the compressed block itself
	_, err = bc.file.Write(compressedBlock)
	if err != nil {
		return err
	}

	// @TODO: add 4 byte checksum!!!

	bc.currentBlock += blockSize

	bc.buffer.Reset()
	return nil
}

func (bc *compressedBlockWriter) Close() error {
	// Flush any remaining data in the buffer
	err := bc.flushBuffer()
	if err != nil {
		return err
	}

	return bc.file.Close()
}
