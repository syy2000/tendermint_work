package v0

import (
	"github.com/tendermint/tendermint/types"
)

//  ================== Partition and Reap ==================================

func (mem *CListMempool) ReapBlocks(n int) (int, [][]types.RawTx) {
	mem.partition_lock.Lock()
	defer mem.partition_lock.Unlock()

	// TODO
	// Choose exactly 1 block, instead of all blocks
	// This will change inputs and outputs

	n, outTxNodeSets := mem.partitionResult.ReapBlocks(n)
	out := make([][]types.RawTx, n)

	for i, txs := range outTxNodeSets {
		tmp := make([]types.RawTx, len(txs))
		for j, txNode := range txs {
			tx := txNode.(*mempoolTx)
			tmp[j] = tx.tx.GetTx()
		}
		out[i] = tmp
	}
	return n, out
}
