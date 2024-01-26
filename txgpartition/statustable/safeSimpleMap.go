package statustable

import (
	"bytes"
	"sort"
	"sync"

	"github.com/tendermint/tendermint/crypto/tmhash"
)

type SafeSimpleMap struct {
	queryTable sync.Map
}

var _ Table = (*SimpleMap)(nil)

func NewSafeSimpleMap() *SafeSimpleMap {
	return &SafeSimpleMap{
		queryTable: sync.Map{},
	}
}

func (m *SafeSimpleMap) clear() {
	m.queryTable = sync.Map{}
}
func (m *SafeSimpleMap) get(key string) (Stringer, bool) {
	out, ok := m.queryTable.Load(key)
	if !ok {
		return nil, false
	} else if value, ok := out.(Stringer); !ok {
		return nil, false
	} else {
		return value, true
	}
}
func (m *SafeSimpleMap) set(key string, value Stringer) bool {
	m.queryTable.Store(key, value)
	return true
}
func (m *SafeSimpleMap) hash() []byte {
	var (
		b      bytes.Buffer
		keySet []string = make([]string, 0)
	)
	queryKey := func(key any, value any) bool {
		if k, ok := key.(string); !ok {
			panic("this should never happen! Have you changed safeSimpleMap.go?")
		} else {
			keySet = append(keySet, k)
		}
		return true
	}
	m.queryTable.Range(queryKey)
	sort.Strings(keySet)
	for _, key := range keySet {
		n, ok := m.get(key)
		if !ok {
			panic("this should never happen! Maybe you have been clearing this mapping table while calculating its hash, which shuoldn't happen when everything is right!")
		}
		valueHash := tmhash.Sum([]byte(n.String()))
		nodeHash := tmhash.Sum(join2Bytes(valueHash, []byte(key)))
		if _, err := b.Write(nodeHash); err != nil {
			panic(err)
		}
	}
	return tmhash.Sum(b.Bytes())
}
