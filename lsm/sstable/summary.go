package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/core/ioutils"
)

func ReadSummaryFile[P core.PartKeyTypes](file io.Reader, ss *SSTable[P]) error {
	var err error

	// 1st part - (ascii) metadata
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "------" {
			break
		}

		spl := strings.Split(line, "=")
		sv, err := strconv.ParseFloat(spl[1], 32)

		if err != nil {
			if err == io.EOF {
				break
			}

			continue
		}
		ss.Statistics[spl[0]] = float32(sv)
	}

	// 2nd part - (binary) key ranges for .idx
	for {
		sum := IndexSummary[P]{}

		err = ioutils.ReadDynamicValue[uint32](file, &sum.PartKey)
		if err != nil {
			if err == io.EOF {
				// EOF really should only occur here
				break
			}
			return err
		}

		err = ioutils.ReadDynamicValue[uint32](file, &sum.MaxKey)
		if err != nil {
			return err
		}

		err = binary.Read(file, binary.BigEndian, sum.MinBlockOffset)
		if err != nil {
			return err
		}

		err = binary.Read(file, binary.BigEndian, sum.MaxBlockOffset)
		if err != nil {
			return err
		}

		ss.summaries = append(ss.summaries, sum)
	}

	return nil
}

type summaryWriter[P core.PartKeyTypes] struct {
	filename       string
	file           *os.File
	writer         *bufio.Writer
	currentSummary *IndexSummary[P]
}

func NewSummaryWriter[P core.PartKeyTypes](filename string) (*summaryWriter[P], error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(file)

	return &summaryWriter[P]{
		filename: filename,
		file:     file,
		writer:   w,
	}, nil
}

func (s summaryWriter[P]) WriteMetadata(metadata map[string]string, statistics map[string]int) error {
	s.writer.WriteString("------")

	return nil
}

func (s *summaryWriter[P]) StartRegion(id int, keyLength uint32, key P, indexFileOffset uint32) (*IndexSummary[P], error) {
	s.currentSummary = &IndexSummary[P]{
		Id:              int16(id),
		PartKey:         key,
		IndexFileOffset: indexFileOffset,
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, keyLength)
	binary.Write(buf, binary.BigEndian, s.currentSummary.PartKey)
	binary.Write(buf, binary.BigEndian, s.currentSummary.IndexFileOffset)

	_, err := s.file.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return s.currentSummary, nil
}

func (s *summaryWriter[P]) StopRegion(key P, indexBlockOffset uint32) (*IndexSummary[P], error) {

	sum := s.currentSummary
	s.currentSummary = nil

	return sum, nil
}

func (s summaryWriter[P]) Empty() bool {
	return s.currentSummary == nil
}

func (s summaryWriter[P]) AssertBlocksetOK() error {
	// if s.currentSummary.MaxBlockOffset != 0 {
	// 	return fmt.Errorf("WTF ERROR: @@__@ THIS SHOULD BE 0: %s", s.currentSummary.MaxBlockOffset)
	// }

	return nil
}

func (s *summaryWriter[P]) Close() error {
	err := s.writer.Flush()
	if err != nil {
		return err
	}

	return s.file.Close()
}
