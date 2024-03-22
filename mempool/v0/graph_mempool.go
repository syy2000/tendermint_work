package v0

/*
modified : 2023/8/11 AM
name : donghao

modified : 2023/8/15 AM
name : donghao
*/
import (
	"sort"

	txp "github.com/tendermint/tendermint/txgpartition"
)

var _ txp.TxGraph = (*CListMempool)(nil)

func MustMempoolTx(n txp.TxNode) *mempoolTx {
	if o, ok := n.(*mempoolTx); ok {
		return o
	} else if n == nil {
		return nil
	} else {
		panic("this should not happen! a non mempoolTx is delivered by txgpartition!")
	}
}

func (mmp *CListMempool) IsBlockNode(n txp.TxNode) bool {
	u := MustMempoolTx(n)
	return u.isBlock
}

func (mmp *CListMempool) InDegree(n txp.TxNode) int {
	u := MustMempoolTx(n)
	return u.inDegree
}
func (mmp *CListMempool) OutDegree(n txp.TxNode) int {
	u := MustMempoolTx(n)
	return u.outDegree
}
func (mmp *CListMempool) DecOutDegree(n txp.TxNode) {
	u := MustMempoolTx(n)
	u.outDegree -= 1
}
func (mmp *CListMempool) Visit(n txp.TxNode) {
	// DO NOTHING
}
func (mmp *CListMempool) Visited(n txp.TxNode) bool {
	return false
}
func (mmp *CListMempool) NodeIndex(n txp.TxNode) int64 {
	u := MustMempoolTx(n)
	return u.tx.TxId
}
func (mmp *CListMempool) TxNodeNum() int {
	return mmp.txNodeNum
}
func (mmp *CListMempool) BlockNodeNum() int {
	return mmp.blockNodeNum
}

func (mmp *CListMempool) QueryFather(n txp.TxNode) []txp.TxNode {
	u := MustMempoolTx(n)
	return u.parentTxs
}

func (mmp *CListMempool) QueryNodeChild(n txp.TxNode) []txp.TxNode {
	u := MustMempoolTx(n)
	return u.childTxs
}

func (mmp *CListMempool) FindZeroOutdegree() []txp.TxNode {
	out := make([]txp.TxNode, 0)
	for _, n := range mmp.workspace {
		if n.outDegree == 0 {
			out = append(out, n)
		}
	}
	for _, n := range mmp.blockNodes {
		out = append(out, n)
	}
	return out
}

// only for test

func (mmp *CListMempool) edgeNum() int {
	u := 0
	for _, tx := range mmp.workspace {
		u += len(tx.parentTxs)
	}
	return u
}

func (mmp *CListMempool) maxDep() int {
	u := -1
	for _, tx := range mmp.workspace {
		if t := len(tx.parentTxs); t > u {
			u = t
		}
	}
	return u
}

func (mmp *CListMempool) midDep() int {
	u := make([]int, len(mmp.workspace))
	for i, tx := range mmp.workspace {
		u[i] = len(tx.parentTxs)
	}
	toSort := sort.IntSlice(u)
	sort.Sort(toSort)
	return toSort[toSort.Len()/2]
}
