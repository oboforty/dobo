package lsm

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/oboforty/dobo/lsm/sstable"
	"github.com/oboforty/dobo/lsm/utils"
)

func (t *LSMTreeTable[P]) loadSSTables() error {
	dbpath := filepath.Join(t.cfg.SSTable.BasePath, t.TableName)

	// ensure directories
	if err := utils.EnsurePath(dbpath); err != nil {
		return err
	}

	// load relevant tables
	files, err := os.ReadDir(dbpath)
	if err != nil {
		log.Fatal(err)
	}
	r, _ := regexp.Compile("^(.*)-?([0-9]+).dat$")

	for _, file := range files {
		match := r.FindStringSubmatch(file.Name())
		if match == nil {
			println("Skipping ", file.Name())
			continue
		}

		genId, _ := strconv.Atoi(match[2])
		// tableName := match[1]

		sst := sstable.New[P](
			&t.cfg.SSTable,
			t.TableName,
			t.partKeyTypeInfo,
			genId,
		)
		sst.LoadFromDisc()

		t.SSTables = append(t.SSTables, sst)
	}

	return nil
}
