package randomtransaction

import (
	"fmt"
	"testing"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
)

const (
	_block_num = 15
	_tx_num    = 60000
	_split_num = 10
)

func TestTransaction(t *testing.T) {
	s := NewAddressMap(_block_num, _tx_num)
	s.Init()
	fmt.Println("construct time", s.ConstructTime())

	// Partitioning with TOPSort + DFS
	start := time.Now()
	partitioning, cm, txMap := txgpartition.Init_Partitioning(s, _split_num)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))

	// Simple Move
	start = time.Now()
	_, cm, _ = txgpartition.SimpleMove(s, _split_num, 0.25, partitioning, cm, txMap)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))

}
