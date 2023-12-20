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

func (mmp *CListMempool) countComponent() (map[int64][]int64, map[int64]int64, int64) {
	var count int64
	count = 0
	visit := make(map[int64]bool)
	componentMap := make(map[int64][]int64)
	weightMap := make(map[int64]int64)
	for _, tx := range mmp.workspace {
		if !visit[tx.ID()] { // 开启一个新的连通分量
			visit[tx.ID()] = true
			component := []int64{}
			component = append(component, tx.ID())
			var weight int64
			weight = mmp.workspace[tx.ID()].weight
			count += 1
			component, weight = mmp.dfs(tx, visit, component, weight)
			componentMap[count] = component
			weightMap[count] = weight
		}
	}
	return componentMap, weightMap, count
}
func (mmp *CListMempool) dfs(tx txp.TxNode, visit map[int64]bool, component []int64, weight int64) ([]int64, int64) {
	for _, t := range mmp.QueryNodeChild(tx) {
		if !visit[t.ID()] {
			visit[t.ID()] = true
			component = append(component, t.ID())
			weight += mmp.workspace[t.ID()].weight
			component, weight = mmp.dfs(t, visit, component, weight)
		}
	}
	for _, t := range mmp.QueryFather(tx) {
		if !visit[t.ID()] {
			visit[t.ID()] = true
			component = append(component, t.ID())
			weight += mmp.workspace[t.ID()].weight
			component, weight = mmp.dfs(t, visit, component, weight)
		}
	}
	return component, weight
}

func (mmp *CListMempool) countWeight() int64 {
	var weight int64
	for _, tx := range mmp.workspace {
		weight += tx.weight
	}
	return weight
}
