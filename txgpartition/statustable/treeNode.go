package statustable

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
)

var (
	nilHandler = []byte("0")
	nilHash    = tmhash.Sum(nilHandler)
)

type Node interface {
	hash() []byte
}

type BranchNode struct {
	childs [17]Node
}
type ExtendNode struct {
	child Node
	path  []byte
}
type ValueNode struct {
	value Stringer
}

func (branchNode *BranchNode) hash() []byte {
	var hashHandler = []byte("b")
	for _, child := range branchNode.childs {
		if child == nil {
			hashHandler = join2Bytes(hashHandler, nilHandler)
		} else {
			hashHandler = join2Bytes(hashHandler, child.hash())
		}
	}
	return tmhash.Sum(hashHandler)
}
func (extendNode *ExtendNode) hash() []byte {
	var hashHandler = []byte("e")
	if extendNode.child == nil {
		hashHandler = join2Bytes(hashHandler, nilHash)
		hashHandler = join2Bytes(hashHandler, extendNode.path)
	} else {
		hashHandler = join2Bytes(hashHandler, extendNode.child.hash())
		hashHandler = join2Bytes(hashHandler, extendNode.path)
	}
	return tmhash.Sum(hashHandler)
}
func (valueNode *ValueNode) hash() []byte {
	return tmhash.Sum([]byte("v" + valueNode.value.String()))
}
