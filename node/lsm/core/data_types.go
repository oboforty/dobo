package core

import "cmp"

type FindStatus = uint8

const (
	NOT_FOUND FindStatus = iota
	FOUND_AT_MEM
	FOUND_AT_BLOOM
	FOUND_AT_SS
)

type ItemWrite[K cmp.Ordered] struct {
	PartKey K
	SortKey any
	Value   []byte
}

type ItemQuery[K cmp.Ordered] struct {
	PartKey K
	SortKey any
	Value   []byte

	// todo: put these into metadata? or we'll put them at api json lvl?
	FoundIn      FindStatus
	FoundSSLevel int8
	FoundSSIdx   uint32
}

type DynamicValueSerialization string

const (
	// Values are stored as json in byte format. Keys are prepended
	STORE_TYPE_JSONB = "jsonb"
	// Values are stored as json lines file, with added __KEY and __SRT fields
	STORE_TYPE_JSONL = "jsonl"
	// Values are serialized as gob
	STORE_TYPE_GOB = "gob"
	// Values are serialized in their binary formats
	STORE_TYPE_BINARY = "binary"
)
