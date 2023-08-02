package statustable

func mptGet(n0 Node, key []byte) (Stringer, bool) {
	if n0 == nil {
		return nil, false
	}
	switch n := n0.(type) {
	case *BranchNode:
		if len(key) == 0 {
			return mptGet(n.childs[16], key)
		} else {
			if idx := PrefixIndex(key[0]); idx < 0 {
				return nil, false
			} else {
				return mptGet(n.childs[idx], key[1:])
			}
		}
	case *ExtendNode:
		if commonLen := LonggestCommonPrefix(n.path, key); commonLen != len(n.path) {
			return nil, false
		} else {
			return mptGet(n.child, key[commonLen:])
		}
	case *ValueNode:
		if len(key) > 0 {
			return nil, false
		} else {
			return n.value, true
		}
	default:
		return nil, false
	}
}

func mptSet(n0 Node, key []byte, value Stringer) (Node, bool) {
	if n0 == nil {
		return NewKeyPath(key, value), true
	}
	switch n := n0.(type) {
	case *ValueNode:
		if len(key) == 0 {
			n.value = value
			return n, true
		} else {
			idx := PrefixIndex(key[0])
			if idx < 0 {
				return n, false
			} else {
				branchNode := BranchNode{childs: [17]Node{}}
				branchNode.childs[16] = n
				branchNode.childs[idx] = NewKeyPath(key[1:], value)
				return &branchNode, true
			}
		}
	case *BranchNode:
		if len(key) == 0 {
			n.childs[16] = &ValueNode{value: value}
			return n, true
		} else {
			idx := PrefixIndex(key[0])
			if idx < 0 {
				return n, false
			} else {
				if rplc, ok := mptSet(n.childs[idx], key[1:], value); !ok {
					return n, false
				} else {
					n.childs[idx] = rplc
					return n, true
				}
			}
		}
	case *ExtendNode:
		if len(key) == 0 {
			branchNode := BranchNode{childs: [17]Node{}}
			branchNode.childs[16] = &ValueNode{value}
			branchNode.childs[PrefixIndex(n.path[0])] = QuitedExtendNode(n, 1)
			return &branchNode, true
		} else if commonLen := LonggestCommonPrefix(n.path, key); commonLen == 0 {
			branchNode := BranchNode{childs: [17]Node{}}
			idxExtendNode := PrefixIndex(n.path[0])
			idxKey := PrefixIndex(key[0])
			if idxKey < 0 || idxExtendNode < 0 {
				return n, false
			} else {
				branchNode.childs[idxExtendNode] = QuitedExtendNode(n, 1)
				branchNode.childs[idxKey] = NewKeyPath(key[1:], value)
			}
			return &branchNode, true
		} else {
			extendNode := ExtendNode{
				path:  n.path[:commonLen],
				child: QuitedExtendNode(n, commonLen),
			}
			if rplc, ok := mptSet(extendNode.child, key[:commonLen], value); !ok {
				return n, false
			} else {
				extendNode.child = rplc
				return &extendNode, true
			}
		}
	default:
		return n0, false
	}
}

//=====================================================================

func NewKeyPath(key []byte, value Stringer) Node {
	valueNode := &ValueNode{value: value}
	if len(key) == 0 {
		return valueNode
	} else {
		return &ExtendNode{
			child: valueNode,
			path:  key,
		}
	}
}

func QuitedExtendNode(n *ExtendNode, K int) Node {
	if K <= 0 {
		return n
	} else if len(n.path) <= K {
		return n.child
	} else {
		return &ExtendNode{
			path:  n.path[K:],
			child: n.child,
		}
	}
}
