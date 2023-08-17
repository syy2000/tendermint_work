package txgpartition

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
	}
)

func NewTransactionGraphPartitioner(mode string, g TxGraph) *TransactionGraphPartitioner {
	switch {
	case mode == "normal&sm":
		return &TransactionGraphPartitioner{
			graph: g,
			head:  Init_Partitioning,
			path: []LocalSearchFunc{
				SimpleMove,
			},
		}
	case mode == "normal&am":
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
	}
}

// ================ Result =================================
func (p *TransactionGraphPartitionResult) ReapBlocks(n int) (int, [][]TxNode) {
	n, outID := p.colorMap.ReapBlocks(n)
	out := make([][]TxNode, n)
	for i, id := range outID {
		out[i] = p.txMap.Get(id)
	}
	return n, out
}
