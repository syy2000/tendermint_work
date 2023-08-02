package statustable

import "github.com/tendermint/tendermint/crypto/tmhash"

type MPT struct {
	root    Node
	useHash bool
}

var _ Table = (*MPT)(nil)

func NewMPT() *MPT {
	return &MPT{
		root:    nil,
		useHash: false,
	}
}

func (m *MPT) convert2Hex(i string) []byte {
	if m.useHash {
		return Byte2Hex(tmhash.Sum([]byte(i)))
	}
	return Str2Hex(i)
}
func (m *MPT) hash() []byte {
	if m.root == nil {
		return nil
	} else {
		return m.root.hash()
	}
}
func (m *MPT) get(key string) (Stringer, bool) {
	return mptGet(m.root, m.convert2Hex(key))
}
func (m *MPT) set(key string, value Stringer) bool {
	if rplc, ok := mptSet(m.root, m.convert2Hex(key), value); ok {
		m.root = rplc
		return true
	} else {
		return false
	}
}
func (m *MPT) clear() {
	m.root = nil
}

// ============== Option ==============================
func MPTUseHash(t Table) {
	if o, ok := t.(*MPT); ok {
		o.useHash = true
	}
}
