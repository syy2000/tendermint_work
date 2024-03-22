package v0

import (
	"fmt"
	// "math"
	// "math/rand"
	// "time"
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

func (mmp *CListMempool) countWeight() int64 {
	var weight int64
	for _, tx := range mmp.workspace {
		weight += tx.weight
	}
	return weight
}
