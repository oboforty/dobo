package bloom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bits-and-blooms/bloom/v3"
)

type CfgBloomFilter struct {
	Bits          uint32
	HashFunctions uint32

	MaxItems          uint32  `toml:"max_items" json:"max_items"`
	FalsePositiveRate float64 `toml:"false_positive_rate" json:"false_positive_rate"`
}

// type BloomFilterComparator func(bloom *bloom.BloomFilter, val interface{}) bool

type bloomFilter struct {
	bloom             *bloom.BloomFilter
	maxItems          uint32
	falsePositiveRate float64
}

func New(cfg CfgBloomFilter) *bloomFilter {
	bf := &bloomFilter{
		maxItems:          cfg.MaxItems,
		falsePositiveRate: cfg.FalsePositiveRate,
	}

	if cfg.MaxItems > 0 {
		if cfg.Bits > 0 || cfg.HashFunctions > 0 {
			slog.Warn("[Bloom] redundant config: either define maxItems+FalsePositiveRates OR Bits+HashFunctions in config!")
		}

		// convenience params
		bf.bloom = bloom.NewWithEstimates(uint(bf.maxItems), bf.falsePositiveRate)
	} else if cfg.Bits > 0 {
		bf.bloom = bloom.New(uint(cfg.Bits), uint(cfg.HashFunctions))
	} else {
		// ignore creation
	}

	return bf
}

func (b *bloomFilter) Test(val interface{}) (bool, error) {
	buf := new(bytes.Buffer)

	// @TODO: handle string and byte keys! binary uses Reflect!
	err := binary.Write(buf, binary.BigEndian, val)

	if err != nil {
		return false, err
	}

	return b.bloom.Test(buf.Bytes()), nil
}

func (b *bloomFilter) Add(val interface{}) error {
	buf := new(bytes.Buffer)

	// @TODO: handle string and byte keys! binary uses Reflect!
	err := binary.Write(buf, binary.BigEndian, val)

	if err != nil {
		return err
	}

	b.bloom.Add(buf.Bytes())

	return nil
}

func (b *bloomFilter) FalsePositiveRate() float64 {
	return b.falsePositiveRate
}

func (b *bloomFilter) MaxItems() uint32 {
	return uint32(b.maxItems)
}

func (b1 *bloomFilter) Merge(b2 interface{}) error {
	bf2, ok := b2.(*bloomFilter)
	if !ok {
		return fmt.Errorf("invalid BF type for merge")
	}

	return b1.bloom.Merge(bf2.bloom)
}

func (b *bloomFilter) WriteToDisc(filename string) error {
	data, err := b.bloom.MarshalJSON()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.BigEndian, data)

	return err
}

func (b *bloomFilter) LoadFromDisc(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = b.bloom.UnmarshalJSON(data)

	return err
}
