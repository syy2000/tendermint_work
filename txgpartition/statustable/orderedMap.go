package statustable

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto/tmhash"
)

type StringerNode struct {
	key   string
	value Stringer
	next  *StringerNode
}

type OrderedMap struct {
	queryTable map[string]Stringer
	hashList   *StringerNode
}

var _ Table = (*OrderedMap)(nil)

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		queryTable: make(map[string]Stringer),
		hashList:   nil,
	}
}

func (om *OrderedMap) set(key string, value Stringer) bool {
	om.queryTable[key] = value
	u := &StringerNode{
		key:   key,
		value: value,
		next:  om.hashList,
	}
	om.hashList = u
	return true
}

func (om *OrderedMap) get(key string) (Stringer, bool) {
	u, ok := om.queryTable[key]
	return u, ok
}

func (om *OrderedMap) clear() {
	om.hashList = nil
	om.queryTable = map[string]Stringer{}
}

func (om *OrderedMap) hash() []byte {
	n := om.hashList
	var b bytes.Buffer
	for n != nil {
		valueHash := tmhash.Sum([]byte(n.value.String()))
		nodeHash := tmhash.Sum(join2Bytes(valueHash, []byte(n.key)))
		if _, err := b.Write(nodeHash); err != nil {
			panic(err)
		}
		n = n.next
	}
	return tmhash.Sum(b.Bytes())
}
