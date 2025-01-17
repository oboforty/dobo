package sstable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (ss *SSTable) TableExists() bool {
	var fn = fmt.Sprintf("t_%s_%s", ss.ParentTableName, ss.Idx)

	_, err := os.Stat(filepath.Join(TOIGHT, fn+".stats"))

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (ss *SSTable) ReadOffsetIndex(partKey interface{}) {
	// file, err := os.Open(filePath)

	// @TODO: implement binary search on .idx & .db

	// @TODO: read Cassandra's summary table -- wtf does it contain anyway?
}

func (ss *SSTable) createTable() {
	// Creates internal files

	var fn = fmt.Sprintf("t_%s_%d", ss.ParentTableName, ss.Idx)
	createEmptyFile(filepath.Join(TOIGHT, fn+".db"))
	createEmptyFile(filepath.Join(TOIGHT, fn+".idx"))
	createEmptyFile(filepath.Join(TOIGHT, fn+".bf"))
	createEmptyFile(filepath.Join(TOIGHT, fn+".stats"))
	createEmptyFile(filepath.Join(TOIGHT, fn+".crc"))
}
