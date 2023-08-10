package randomsquare

import (
	"fmt"
	"testing"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
)

const (
	GRAPHSIZE  = 12000
	_split_num = 20
)

func TestSquare(t *testing.T) {
	s := NewRandomSquare(GRAPHSIZE)
	s.RandomInit(0.8)
	fmt.Println("ok")
	start := time.Now()
	partitioning, cm, txMap := txgpartition.Init_Partitioning(s, _split_num, 0.0)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))
	fmt.Println("cut : ", txgpartition.CalculatePartitioningQualityByInnerPartitioningEdge(cm))

	// Simple Move
	start = time.Now()
	partitioning, cm, txMap = txgpartition.SimpleMove(s, _split_num, 0.4, partitioning, cm, txMap)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))

	// Advanced Move
	start = time.Now()
	_, cm, _ = txgpartition.AdvancedMove(s, _split_num, 0.4, partitioning, cm, txMap)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))
}
