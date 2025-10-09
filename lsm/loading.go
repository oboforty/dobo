package lsm

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pelletier/go-toml/v2"

	"github.com/oboforty/dobo/lsm/bloom"
	"github.com/oboforty/dobo/lsm/core"
	"github.com/oboforty/dobo/lsm/memtable"
	"github.com/oboforty/dobo/lsm/sstable"
)

type CfgTable struct {
	Name     string               `toml:"name" json:"name"`
	KeyType  core.DataType        `toml:"key_type" json:"key_type"`
	MemTable memtable.CfgMemtable `toml:"memtable" json:"memtable"`
	SSTable  sstable.CfgSSTable   `toml:"sstable" json:"sstable"`
}

var tableDefaults = CfgTable{
	KeyType: "int64",
	MemTable: memtable.CfgMemtable{
		Type:        memtable.MEMTYPE_REDBLACK,
		MaxByteSize: 65536,
	},
	SSTable: sstable.CfgSSTable{
		CompressionBlockSize: 16384,
		MinIdxInterval:       128,
		MaxIdxInterval:       4096,
		MaxSumSize:           5242880,

		BloomFilter: bloom.CfgBloomFilter{
			FalsePositiveRate: 0.1,
		},
	},
}

func (cfg *CfgTable) ApplyDefaults() {
	// Global Table Defaults
	if cfg.KeyType == "" {
		cfg.KeyType = tableDefaults.KeyType
	}

	// MemTable Defaults
	if cfg.MemTable.Type == "" {
		cfg.MemTable.Type = tableDefaults.MemTable.Type
	}
	if cfg.MemTable.MaxByteSize == 0 {
		cfg.MemTable.MaxByteSize = tableDefaults.MemTable.MaxByteSize
	}

	// SSTable defaults
	if cfg.SSTable.CompressionBlockSize == 0 {
		cfg.SSTable.CompressionBlockSize = tableDefaults.SSTable.CompressionBlockSize
	}
	if cfg.SSTable.MinIdxInterval == 0 {
		cfg.SSTable.MinIdxInterval = tableDefaults.SSTable.MinIdxInterval
	}
	if cfg.SSTable.MaxIdxInterval == 0 {
		cfg.SSTable.MaxIdxInterval = tableDefaults.SSTable.MaxIdxInterval
	}

	// BF defaults
	if cfg.SSTable.BloomFilter.FalsePositiveRate == 0 {
		cfg.SSTable.BloomFilter.FalsePositiveRate = tableDefaults.SSTable.BloomFilter.FalsePositiveRate
	}
}

func (cfg *CfgTable) WriteToDisc() error {
	tablePath := filepath.Join(cfg.SSTable.DBPath, cfg.Name, "table.toml")

	err := core.EnsurePath(filepath.Dir(tablePath))
	if err != nil {
		slog.Error(fmt.Sprintf("[%s] config error: couldn't create directories: %s (%s)", cfg.Name, err, tablePath))
		return err
	}

	data, err := toml.Marshal(cfg)

	if err != nil {
		slog.Error(fmt.Sprintf("[%s] config toml serialize error: %s (%s)", cfg.Name, err, tablePath))
		return err
	}

	err = os.WriteFile(tablePath, data, 0644)

	if err != nil {
		slog.Error(fmt.Sprintf("[%s] config write error: %s (%s)", cfg.Name, err, tablePath))
		return err
	}

	return nil
}

func ReadTableConfig(dbPath string) (*CfgTable, error) {
	tablePath := filepath.Join(dbPath, "table.toml")
	cfgContent, err := os.ReadFile(tablePath)

	if err != nil {
		slog.Error(fmt.Sprintf("[Cfg] config file not found: %s (%s)", err, tablePath))
		return nil, err
	}

	cfgiTable := &CfgTable{}

	err = toml.Unmarshal(cfgContent, cfgiTable)
	if err != nil {
		slog.Error(fmt.Sprintf("[%s] config file not found: %s (%s)", err, tablePath))
		return nil, err
	}

	return cfgiTable, nil
}

func (t *LSMTreeTable[P]) loadSSTables() error {
	dbPath := filepath.Join(t.cfg.SSTable.DBPath, t.TableName())

	// load relevant tables
	files, err := os.ReadDir(dbPath)
	if err != nil {
		return err
	}
	r, _ := regexp.Compile("^(.*)-?([0-9]+).dat$")

	for _, file := range files {
		match := r.FindStringSubmatch(file.Name())
		if match == nil {
			continue
		}

		genId, _ := strconv.Atoi(match[2])
		sst := sstable.New[P](
			&t.cfg.SSTable,
			t.cfg.Name,
			t.partKeyTypeInfo,
			genId,
		)

		slog.Info(fmt.Sprintf("[%s] loading from %s", t.cfg.Name, sst.FileBase()))
		sst.LoadFromDisc()

		t.SSTables = append(t.SSTables, sst)
	}

	return nil
}

type LSMTreeTableInterface interface {
	TableName() string
	TableInfo() *CfgTable
	// Get(partKey P) *core.ItemQuery[P]

	FlushMemToDisc() error
	// @TODO: add more useful funcs to this interface
}

func NewFromDisc(dbPath string) (LSMTreeTableInterface, error) {
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
