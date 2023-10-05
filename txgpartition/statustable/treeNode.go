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
	var hashHandler []byte
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
	if extendNode.child == nil {
		return tmhash.Sum(join2Bytes(nilHash, extendNode.path))
	} else {
		return tmhash.Sum(join2Bytes(extendNode.child.hash(), extendNode.path))
	}
}
func (valueNode *ValueNode) hash() []byte {
	return tmhash.Sum([]byte("1" + valueNode.value.String()))
}
