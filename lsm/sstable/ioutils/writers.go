package ioutils

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
)

type BlockWriter interface {
	io.Writer
	Close() error
	GetOffsets() (blockOffset uint32, interBlockOffset uint32)
}

func NewBlockWriter(filename string, blockSize uint32, compressMethod bool) (BlockWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	if compressMethod {
		// @TODO: also add buffered IO for this!
		return &compressedBlockWriter{
			blockSize: int(blockSize),
			// filename:  filename,
			file:   file,
			buffer: bytes.Buffer{},
		}, nil
	} else {
		w := bufio.NewWriter(file)

		return &blockWriter{
			blockSize: int(blockSize),
			// filename:  filename,
			file:      file,
			bufWriter: w,
		}, nil
	}
}

type blockWriter struct {
	// filename  string
	file      *os.File
	bufWriter *bufio.Writer

	blockSize        int
	blockOffset      uint32
	interBlockOffset uint32
}

func (bw *blockWriter) Write(data []byte) (int, error) {

	nb, err := bw.bufWriter.Write(data)

	if err == nil {
		bw.interBlockOffset += uint32(nb)

		if bw.interBlockOffset > uint32(bw.blockSize) {
			bw.blockOffset += bw.interBlockOffset
			bw.interBlockOffset = 0
		}
	}

	return nb, err
}

func (bc *blockWriter) GetOffsets() (uint32, uint32) {
	return bc.blockOffset, bc.interBlockOffset
}

func (bw *blockWriter) Close() error {
	err := bw.bufWriter.Flush()

	if err != nil {
		return err
	}

	return bw.file.Close()
}

// compressedBlockWriter handles writing compressed blocks of data to a file with buffering
type compressedBlockWriter struct {
	// filename string
	file *os.File

	// compressMethod CompressionAlgorithm
	buffer           bytes.Buffer
	blockSize        int
	currentBlock     uint32
	interBlockOffset uint32
}

func (bc *compressedBlockWriter) Write(data []byte) (int, error) {
	n, err := bc.buffer.Write(data)

	if err == nil {
		bc.interBlockOffset += uint32(n)

		if bc.buffer.Len() >= bc.blockSize {
			err = bc.flushBuffer()
		}
	}

	return n, err
}

func (bc *compressedBlockWriter) flushBuffer() error {
	if bc.buffer.Len() == 0 {
		return nil
	}

	compressedBlock, err := compressBlock(bc.buffer.Bytes())
	if err != nil {
		return err
	}

	// Write the size of the compressed block first
	blockSize := uint32(len(compressedBlock))
	err = binary.Write(bc.file, binary.BigEndian, blockSize)
	if err != nil {
		return err
	}

	// Write the compressed block itself
	_, err = bc.file.Write(compressedBlock)
	if err != nil {
		return err
	}

	// @TODO: add 4 byte checksum!!!

	bc.currentBlock += blockSize + 4
	bc.interBlockOffset = 0
	bc.buffer.Reset()

	return nil
}

func (bc *compressedBlockWriter) GetOffsets() (uint32, uint32) {
	return bc.currentBlock, bc.interBlockOffset
}

func (bc *compressedBlockWriter) Close() error {
	// Flush any remaining data in the buffer
	err := bc.flushBuffer()
	if err != nil {
		return err
	}

	return bc.file.Close()
}

func compressBlock(data []byte) ([]byte, error) {
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
