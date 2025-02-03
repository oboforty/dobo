package sstable

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/oboforty/dobo/lsm/sstable/ioutils"
)

func (ss *SSTable[P]) LoadFromDisc() error {
	file, err := os.Open(ss.tablePath + ".dat")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// load bloom filter
	ss.bloom.LoadFromDisc(ss.tablePath + ".bf")

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
		sum := &IndexSummary[P]{
			PartKeyTypeInfo: &ss.partKeyTypeInfo,
		}

		err = ioutils.ReadDynamicValue[uint32](file, &sum.MinKey)
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
	}

	return nil
}
