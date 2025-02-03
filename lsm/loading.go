package lsm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
	"github.com/pelletier/go-toml/v2"
)

type CfgTable struct {
	Name     string               `toml:"name"`
	KeyType  core.DataType        `toml:"key_type"`
	MemTable memtable.CfgMemtable `toml:"memtable"`
	SSTable  sstable.CfgSSTable   `toml:"sstable"`
}

var tableDefaults = CfgTable{
	// KeyType
	MemTable: memtable.CfgMemtable{
		Type:        memtable.MEMTYPE_REDBLACK,
		MaxByteSize: 65536,
	},
	SSTable: sstable.CfgSSTable{
		CompressionBlockSize: 16384,
		BloomFilter: sstable.CfgBloomFilter{
			FalsePositiveRate: 0.1,
		},
	},
}

func (cfg *CfgTable) ApplyDefaults() {
	// @TODO: add a defaults.toml at top lvl?
	//				ApplyDefaults(defaults *CfgTable, ...) {

	if cfg.MemTable.Type == "" {
		cfg.MemTable.Type = tableDefaults.MemTable.Type
	}

	if cfg.MemTable.MaxByteSize == 0 {
		cfg.MemTable.MaxByteSize = tableDefaults.MemTable.MaxByteSize
	}

	if cfg.SSTable.CompressionBlockSize == 0 {
		cfg.SSTable.CompressionBlockSize = tableDefaults.SSTable.CompressionBlockSize
	}

	if cfg.SSTable.BloomFilter.FalsePositiveRate == 0 {
		cfg.SSTable.BloomFilter.FalsePositiveRate = tableDefaults.SSTable.BloomFilter.FalsePositiveRate
	}

}

func (t *LSMTreeTable[P]) loadSSTables() error {
	dbPath := filepath.Join(t.cfg.SSTable.DBPath, t.TableName())

	// ensure directories (? is this needed?)
	// if err := utils.EnsurePath(dbPath); err != nil {
	// 	return err
	// }

	// load relevant tables
	files, err := os.ReadDir(dbPath)
	if err != nil {
		return err
	}
	r, _ := regexp.Compile("^(.*)-?([0-9]+).dat$")

	for _, file := range files {
		match := r.FindStringSubmatch(file.Name())
		if match == nil {
			log.Println("[SST] Skipping non-table ", file.Name())
			continue
		}

		genId, _ := strconv.Atoi(match[2])
		// tableName := match[1]

		sst := sstable.New[P](
			&t.cfg.SSTable,
			t.cfg.Name,
			t.partKeyTypeInfo,
			genId,
		)
		sst.LoadFromDisc()

		t.SSTables = append(t.SSTables, sst)
	}

	return nil
}

type LSMTreeTableInterface interface {
	TableName() string

	// @TODO: add more useful funcs to this interface
}

func ReadTableConfig(dbPath string) (*CfgTable, error) {

	cfgContent, err := os.ReadFile(filepath.Join(dbPath, "table.toml"))
	if err != nil {
		log.Fatalf("[Cfg] Unable to find config file at %s", dbPath)
		return nil, err
	}

	cfgiTable := &CfgTable{}

	err = toml.Unmarshal(cfgContent, cfgiTable)
	if err != nil {
		log.Fatalf("[Cfg] parse error: %s", err)
		return nil, err
	}

	return cfgiTable, nil
}

func NewFromDisc(dbPath string) (LSMTreeTableInterface, error) {
	// tablePath := filepath.Join(cfg.DBPath, tableName, strconv.Itoa(id))
	// utils.EnsurePath(filepath.Dir(tablePath))

	// utf8.Valid(
	// scanner := bufio.NewScanner(file)
	// scanner.Scan()
	// scanner.Text()

	cfg, err := ReadTableConfig(dbPath)
	if err != nil {
		return nil, err
	}

	// by default SSTables are stored where table.toml is at
	if cfg.SSTable.DBPath == "" {
		cfg.SSTable.DBPath = dbPath
	}

	var tree LSMTreeTableInterface

	switch cfg.KeyType {
	case core.DTYPE_INT32:
		tree, err = New[int32](cfg)
	case core.DTYPE_INT64:
		tree, err = New[int64](cfg)
	case core.DTYPE_FLOAT32:
		tree, err = New[float32](cfg)
	case core.DTYPE_FLOAT64:
		tree, err = New[float64](cfg)
	case core.DTYPE_BYTES:
		return nil, fmt.Errorf("bytes are not yet supported in this version %s", ":3")
		// tree, err = New[core.ByteSlice](cfg)
	case core.DTYPE_STRING:
		tree, err = New[string](cfg)
	default:
		return nil, fmt.Errorf("invalid data type: %s", cfg.KeyType)
	}

	if err != nil {
		return nil, err
	}

	return tree, nil
}
