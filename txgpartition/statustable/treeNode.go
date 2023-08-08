package statustable

import "github.com/tendermint/tendermint/crypto/tmhash"

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
			hashHandler = append(hashHandler, nilHandler...)
		} else {
			hashHandler = append(hashHandler, child.hash()...)
		}
	}
	return tmhash.Sum(hashHandler)
}
func (extendNode *ExtendNode) hash() []byte {
	if extendNode.child == nil {
		return tmhash.Sum(append(nilHash, extendNode.path...))
	} else {
		return tmhash.Sum(append(extendNode.child.hash(), extendNode.path...))
	}
}
func (valueNode *ValueNode) hash() []byte {
	return tmhash.Sum([]byte("1" + valueNode.value.String()))
}
