package statustable

import (
	"bytes"
	"sort"

	"github.com/tendermint/tendermint/crypto/tmhash"
)

type SimpleMap struct {
	queryTable map[string]Stringer
}

var _ Table = (*SimpleMap)(nil)

func NewSimpleMap() *SimpleMap {
	return &SimpleMap{make(map[string]Stringer)}
}

func (sm *SimpleMap) clear() {
	sm.queryTable = make(map[string]Stringer)
}
func (sm *SimpleMap) get(key string) (Stringer, bool) {
	u, ok := sm.queryTable[key]
	return u, ok
}
func (sm *SimpleMap) set(key string, value Stringer) bool {
	sm.queryTable[key] = value
	return true
}
func (sm *SimpleMap) hash() []byte {
	var (
		b      bytes.Buffer
		keySet []string = make([]string, len(sm.queryTable))
		cnt             = 0
	)
	for key := range sm.queryTable {
		keySet[cnt] = key
		cnt++
	}
	sort.Strings(keySet)
	for _, key := range keySet {
		n := sm.queryTable[key]
		valueHash := tmhash.Sum([]byte(n.String()))
		nodeHash := tmhash.Sum(join2Bytes(valueHash, []byte(key)))
		if _, err := b.Write(nodeHash); err != nil {
			panic(err)
		}
	}
	return tmhash.Sum(b.Bytes())
}
