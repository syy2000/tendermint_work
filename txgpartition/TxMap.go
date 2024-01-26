package txgpartition

type TransactionMap struct {
	nodeMap      [][]TxNode
	blockNum     int
	partitionNum int
	avasize      int
	size         int
}

func NewTransactionMap(b, k, ava int) *TransactionMap {
	return &TransactionMap{
		blockNum:     b,
		partitionNum: k,
		size:         b + k,
		avasize:      ava,
		nodeMap:      make([][]TxNode, b+k),
	}
}

func (t *TransactionMap) Get(idx int) []TxNode {
	return t.nodeMap[idx]
}
func (t *TransactionMap) Append(idx int, tx TxNode) {
	if t.nodeMap[idx] == nil {
		if idx < t.blockNum {
			t.nodeMap[idx] = []TxNode{tx}
			return
		} else {
			t.nodeMap[idx] = append(make([]TxNode, 0, t.avasize), tx)
			return
		}
	} else {
		t.nodeMap[idx] = append(t.nodeMap[idx], tx)
	}
}
func (t *TransactionMap) Size() int {
	return t.size
}
