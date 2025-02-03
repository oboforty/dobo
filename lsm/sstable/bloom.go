package sstable

import (
	"bytes"
	"encoding/binary"
	"io"
	"log/slog"

	"github.com/bits-and-blooms/bloom/v3"
	// "github.com/bits-and-blooms/bloom"
)

type BloomFilterComparator func(bloom *bloom.BloomFilter, val interface{}) bool

type bloomFilter struct {
	// bloom Filter
	// @TODO: put bloom filter into its own struct? + even add interface?
	bloom *bloom.BloomFilter
	// bloomComp BloomFilterComparator
}

type CfgBloomFilter struct {
	Bits          uint32
	HashFunctions uint32

	MaxItems          uint32
	FalsePositiveRate float64
}

func newBloomFilter(cfg CfgBloomFilter) (bf *bloomFilter) {
	if cfg.MaxItems > 0 {
		if cfg.Bits > 0 || cfg.HashFunctions > 0 {
			slog.Warn("[Bloom] redundant config: either define MaxItems+FalsePositiveRates OR Bits+HashFunctions in config!")
		}

		// convenience params
		bf.bloom = bloom.NewWithEstimates(uint(cfg.MaxItems), cfg.FalsePositiveRate)
	} else {
		bf.bloom = bloom.New(uint(cfg.Bits), uint(cfg.HashFunctions))
	}

	// load if exists

	return
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

func (b *bloomFilter) Serialize(w io.Writer) error {
	// data, err := b.bloom.MarshalJSON()
	// if err != nil {
	// 	return err
	// }

	// err := binary.Write(w, binary.BigEndian, val)

	return nil
}
