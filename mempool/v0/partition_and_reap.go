package v0

/*
statement：
本文件中的方法仅适用于第一版
*/

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
	"github.com/tendermint/tendermint/types"
)

// donghao : set block size
func (mem *CListMempool) SetBlockSize(s int) {
	mem.avasize = s
}
func (mem *CListMempool) SetAlpha(alpha float64) {
	mem.alpha = alpha
}

//  ================== Partition and Reap ==================================

// 1. 从工作区中取出至多n个区块（事务集）
// 2. 如果工作区中所有区块都被清空，则
func (mem *CListMempool) ReapBlocks(n int) (int, []types.Txs) {
	// reap_lock ： 保证取区块是串行的
	mem.reap_lock.Lock()
	defer mem.reap_lock.Unlock()

	if mem.partitionResult == nil || mem.partitionResult.Empty() {
		mem.partition_lock.Lock()
		mem.FillWorkspace()
		if mem.txNodeNum == 0 {
			return 1, []types.Txs{nil}
		}
	}

	// partition_lock ： 保证取区块与划分新图是串行的
	mem.partition_lock.Lock()
	defer func() {
		if mem.partitionResult.Empty() {
			// defer mem.partition_lock.Unlock()
			go mem.FillWorkspace()
		} else {
			mem.partition_lock.Unlock()
		}
	}()

	start := time.Now()
	reap_size_cnt := 0
	// TODO Choose exactly 1 block
	n, outTxNodeSets, colors := mem.partitionResult.ReapBlocks(4)
	chosen := rand.Intn(n)

	if n == 0 {
		fmt.Println("============================== ", mem.partitionResult.Empty(), "===============================")
	}
	out := make([]types.Txs, n)

	for i, txs := range outTxNodeSets {
		tmp := make(types.Txs, len(txs))
		reap_size_cnt += len(txs)
		for j, txNode := range txs {
			tx := txNode.(*mempoolTx)
			tmp[j] = types.Tx{OriginTx: tx.tx.GetTx()}
		}
		out[i] = tmp
	}

	// TODO : mapColorToBlockID is only a simple version for test
	mem.mapColorToBlockID(colors)
	mem.logger.Info(fmt.Sprintf("reap %d blocks with avarage block size %d : %s", n, reap_size_cnt/n, time.Since(start)))

	// TODO : Problem
	return n, []types.Txs{out[chosen]}
}

func (mem *CListMempool) FillWorkspace() {
	defer mem.partition_lock.Unlock()

	mem.UpdateBlockStatusMappingTable()

	mem.blockNodeNum = 0
	mem.blockNodes = map[int64]*mempoolTx{}
	mem.workspace, mem.txNodeNum = nil, 0
	mem.blockIDMap = make(map[int]int64)
	mem.txsConflictMap = make(map[string]*txsConflictMapValue)

	mem.moveTxsFromBufferToWorkspace()
	if mem.txNodeNum == 0 {
		return
	}

	start := time.Now()
	fmt.Printf("=============== Partition Size : %d\n", mem.txNodeNum)

	// 事务图生成
	mem.ProcWorkspaceDependency()
	fmt.Println("========== Generate Time : ", time.Since(start), len(mem.workspace))
	start = time.Now()

	// 事务图划分
	mem.SplitWorkspace()
	fmt.Println("========== Partition Time : ", time.Since(start), len(mem.workspace))
	mem.partitionResult.PrintBasic()

	mem.notifiedTxsAvailable = false
	mem.notifyTxsAvailable()
}

func (mem *CListMempool) moveTxsFromBufferToWorkspace() {
	// TODO !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	mem.updateLastTime()
	mem.txNodeNum = len(mem.workspace)
	mem.undo_txs -= mem.txNodeNum
}

func (mem *CListMempool) UpdateBlockStatusMappingTable() {
	for key, dep := range mem.txsConflictMap {
		if len(dep.WL) == 0 {
			continue
		}
		tx := dep.WL[0]
		if tx.isBlock {
			continue
		}
		color, ok := mem.partitionResult.TxNodeColor(tx.ID())
		if !ok {
			continue
		}
		mem.blockStatusMappingTable.Store(key, mem.queryBlockID(color))
	}
}

func (mem *CListMempool) SplitWorkspace() {
	partitioner := txgpartition.NewTransactionGraphPartitioner(
		txgpartition.NORMAL_AM_MODE,
		mem,
	)
	K := int(math.Ceil(float64(mem.txNodeNum) / float64(mem.avasize)))
	alpha := mem.alpha
	mem.partitionResult = partitioner.Partition(K, alpha)
}

func (mem *CListMempool) mapColorToBlockID(colors []int) {
	for _, color := range colors {
		mem.blockIDMap[color] = mem.blockIDCnter
		mem.blockIDCnter++
	}
}
func (mem *CListMempool) queryBlockID(color int) int64 {
	return mem.blockIDMap[color]
}
