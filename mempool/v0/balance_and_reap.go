package v0

<<<<<<< HEAD
import (
	// "fmt"
	// "math"
	// "math/rand"
	// "time"

	// "github.com/tendermint/tendermint/txgpartition"
	// "github.com/tendermint/tendermint/types"
)

func ReapBlocks(componentMap map[int64][]int64, weightMap map[int64]int64, count int64) {

=======
import(
	"fmt"

	"github.com/tendermint/tendermint/types"
)

func(mem *CListMempool) BalanceReapBlocks(componentMap map[int64][]int64, weightMap map[int64]int64, n int64) (int64, []types.Txs){
	if len(componentMap) == 0 || len(weightMap) == 0 {
		fmt.Println("Invalid componentMap or weightMap!")
	}
	mem.reap_lock.Lock()
	defer mem.reap_lock.Unlock()
	out := make([]types.Txs, n)

	// 轮询法
	var(
		i, count    int64
	)
	count = int64(len(componentMap))
	for i=1; i<=count; i++ {
		component := componentMap[i]
		tmp := i%n 
		for _,x := range component {
			mempoolTx := mem.workspace[x]
			out[tmp] = append(out[tmp], types.Tx{OriginTx: mempoolTx.tx.GetTx()})
		}
		
	}
	return n, out
>>>>>>> 002db0e963a7de55957f6893be7d84b708513699
}