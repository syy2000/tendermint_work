package statustable

import (
	"fmt"
	"sync"
)

const (
	UseOrderedMap int8 = iota
	UseSimpleMap
	UseMPTree
	UseSafeSimpleMap
)

var OnlyUseHashOptions = []func(Table){MPTUseHash}

type BlockStatusMappingTable struct {
	table          Table
	blockHashTable sync.Map
}
type (
	blockHeight int64
	blockHash   struct {
		hash []byte
	}
)

func (b blockHeight) String() string {
	return fmt.Sprintf("%d", b)
}

func NewBlockStatusMappingTable(tableType int8, options []func(Table)) *BlockStatusMappingTable {
	var u = BlockStatusMappingTable{blockHashTable: sync.Map{}}
	switch tableType {
	case UseOrderedMap:
		u.table = NewOrderedMap()
	case UseSimpleMap:
		u.table = NewSimpleMap()
	case UseMPTree:
		u.table = NewMPT()
	case UseSafeSimpleMap:
		u.table = NewSafeSimpleMap()
	default:
		panic("unknown type of block status mapping table")
	}
	for _, op := range options {
		op(u.table)
	}
	return &u
}

func (b *BlockStatusMappingTable) Store(key string, value int64) bool {
	return b.table.set(key, blockHeight(value))
}
func (b *BlockStatusMappingTable) Load(key string) (int64, bool) {
	if u, ok := b.table.get(key); ok {
		if tu, ok := u.(blockHeight); ok {
			return int64(tu), true
		}
	}
	return 0, false
}
func (b *BlockStatusMappingTable) Clear() {
	b.table.clear()
	b.blockHashTable = sync.Map{}
}
func (b *BlockStatusMappingTable) Hash() []byte {
	return b.table.hash()
}
func (b *BlockStatusMappingTable) LoadBlockHash(id int64) ([]byte, bool) {
	if out, ok := b.blockHashTable.Load(id); ok {
		if u, ok := out.(*blockHash); ok {
			return u.hash, true
		}
	}
	return nil, false
}
func (b *BlockStatusMappingTable) StoreBlockHash(id int64, hash []byte) bool {
	b.blockHashTable.Store(
		id,
		&blockHash{hash: hash},
	)
	return true
}
