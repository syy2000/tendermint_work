package txgpartition

import "fmt"

const (
	NORMAL_SM_MODE = "normal&sm"
	NORMAL_AM_MODE = "normal&am"
)

type (
	InitPartitionFunc func(TxGraph, int, float64) (*Partitioning, *ColorMap, *TransactionMap)
	LocalSearchFunc   func(TxGraph, int, float64, *Partitioning, *ColorMap, *TransactionMap) (*Partitioning, *ColorMap, *TransactionMap)

	TransactionGraphPartitioner struct {
		graph TxGraph
		head  InitPartitionFunc
		path  []LocalSearchFunc
	}
	TransactionGraphPartitionResult struct {
		partitioning *Partitioning
		colorMap     *ColorMap
		txMap        *TransactionMap
		size         int
	}
)

func NewTransactionGraphPartitioner(mode string, g TxGraph) *TransactionGraphPartitioner {
	switch {
	case mode == NORMAL_SM_MODE:
		return &TransactionGraphPartitioner{
			graph: g,
			head:  Init_Partitioning,
			path: []LocalSearchFunc{
				SimpleMove,
			},
		}
	case mode == NORMAL_AM_MODE:
		return &TransactionGraphPartitioner{
			graph: g,
			head:  Init_Partitioning,
			path: []LocalSearchFunc{
				AdvancedMove,
			},
		}
	default:
		return nil
	}
}

// ===================== Partition ==================================================

func (p *TransactionGraphPartitioner) UpdateGraph(g TxGraph) {
	p.graph = g
}

func (p *TransactionGraphPartitioner) Partition(K int, alpha float64) *TransactionGraphPartitionResult {
	partitioning, cm, tm := p.head(p.graph, K, alpha)
	for _, f := range p.path {
		partitioning, cm, tm = f(p.graph, K, alpha, partitioning, cm, tm)
	}
	return &TransactionGraphPartitionResult{
		partitioning: partitioning,
		colorMap:     cm,
		txMap:        tm,
		size:         cm.numTxBlocks,
	}
}

// ================ Result =================================
func (p *TransactionGraphPartitionResult) ReapBlocks(n int) (int, [][]TxNode, []int) {
	n, outID := p.colorMap.ReapBlocks(n)
	out := make([][]TxNode, n)
	for i, id := range outID {
		out[i] = p.txMap.Get(id)
	}
	p.size -= n
	return n, out, outID
}
func (p *TransactionGraphPartitionResult) Empty() bool {
	return p.size <= 0
}
func (p *TransactionGraphPartitionResult) TxNodeColor(id int64) (int, bool) {
	out, ok := p.partitioning.txColor[id]
	return out, ok
}
func (p *TransactionGraphPartitionResult) PrintBasic() {
	fmt.Println("partition result size : ", p.size, p.txMap.blockNum, p.txMap.partitionNum)
	fmt.Println("partition sets : ", p.colorMap.numTxBlocks)
	fmt.Println("======= TxBlockNum ", p.colorMap.numTxBlocks, "==========")
	fmt.Println("======= PreBlockNum ", p.colorMap.numBlockColor, "==========")
	fmt.Println("======= Cost : ", CalculatePartitioningQualityByColorMap(p.colorMap), "==========")
	fmt.Println("======= Cut : ", CalculatePartitioningQualityByInnerPartitioningEdge(p.colorMap), "==========")
}
