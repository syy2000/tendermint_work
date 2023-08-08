package statustable

import (
	"fmt"
)

const (
	UseOrderedMap int8 = iota
	UseSimpleMap
	UseMPTree
	UseSafeSimpleMap
)

var OnlyUseHashOptions = []func(Table){MPTUseHash}

type BlockStatusMappingTable struct {
	table Table
}
type blockHeight int64

func (b blockHeight) String() string {
	return fmt.Sprintf("%d", b)
}

func NewBlockStatusMappingTable(tableType int8, options []func(Table)) *BlockStatusMappingTable {
	var u BlockStatusMappingTable
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

func (b *BlockStatusMappingTable) Set(key string, value int64) bool {
	return b.table.set(key, blockHeight(value))
}
func (b *BlockStatusMappingTable) Get(key string) (int64, bool) {
	if u, ok := b.table.get(key); ok {
		if tu, ok := u.(blockHeight); ok {
			return int64(tu), true
		}
	}
	return 0, false
}
func (b *BlockStatusMappingTable) Clear() {
	b.table.clear()
}
func (b *BlockStatusMappingTable) Hash() []byte {
	return b.table.hash()
}
