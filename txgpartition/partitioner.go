package txgpartition

type (
	InitPartitionFunc func(TxGraph, int, float64) (map[int64]int, *ColorMap, *TransactionMap)
	LocalSearchFunc   func(TxGraph, int, float64, map[int64]int, *ColorMap, *TransactionMap) (map[int64]int, *ColorMap, *TransactionMap)

	TransactionGraphPartitioner struct {
		graph TxGraph
		head  InitPartitionFunc
		path  []LocalSearchFunc
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

func (p *TransactionGraphPartitioner) UpdateGraph(g TxGraph) {
	p.graph = g
}

func (p *TransactionGraphPartitioner) Partition(K int, alpha float64) (map[int64]int, *ColorMap, *TransactionMap) {
	partitioning, cm, tm := p.head(p.graph, K, alpha)
	for _, f := range p.path {
		partitioning, cm, tm = f(p.graph, K, alpha, partitioning, cm, tm)
	}
	return partitioning, cm, tm
}
