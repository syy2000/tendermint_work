package v0

import (
	"fmt"
	// "math"
	// "math/rand"
	// "time"

	txp "github.com/tendermint/tendermint/txgpartition"
	"github.com/tendermint/tendermint/types"
)

// diploma design
func (mem *CListMempool) BalanceReapBlocks(componentMap map[int64][]int64, weightMap map[int64]int64, n int64) (int64, []types.Txs) {
	if len(componentMap) == 0 || len(weightMap) == 0 {
		fmt.Println("Invalid componentMap or weightMap!")
	}
	mem.reap_lock.Lock()
	defer mem.reap_lock.Unlock()
	out := make([]types.Txs, n)

	// 轮询法
	var (
		i, count int64
	)
	count = int64(len(componentMap))
	// for i = 1; i <= count; i++ {
	// 	component := componentMap[i]
	// 	tmp := i % n
	// 	for _, x := range component {
	// 		mempoolTx := mem.workspace[x]
	// 		out[tmp] = append(out[tmp], types.Tx{OriginTx: mempoolTx.tx.GetTx()})
	// 	}

	// }
	//最小活跃数法(找当前权重和最小的块放入)
	currentWeight := make([]int64, n) // 记录每个预打包块当前的权重
	for i = 1; i <= count; i++ {
		component := componentMap[i]
		tmp := findMinWeightBlock(currentWeight)
		for _, x := range component {
			mempoolTx := mem.workspace[x]
			out[tmp] = append(out[tmp], types.Tx{OriginTx: mempoolTx.tx.GetTx()})
			currentWeight[tmp] += mempoolTx.weight
		}
	}
	return n, out

}
func findMinWeightBlock(currentWeight []int64) int {
	minWeight := currentWeight[0]
	index := 0
	for i := 1; i < len(currentWeight); i++ {
		if currentWeight[i] < minWeight {
			minWeight = currentWeight[i]
			index = i
		}
	}
	return index
}

// diploma design
func (mmp *CListMempool) CountComponent() (map[int64][]int64, map[int64]int64, int64) {
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
