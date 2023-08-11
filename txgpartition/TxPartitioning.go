package txgpartition

type (
	Partitioning struct {
		blockColor, txColor map[int64]int
		graph               TxGraph
	}
)

func NewPartitioning(g TxGraph) *Partitioning {
	return &Partitioning{
		blockColor: make(map[int64]int, g.BlockNodeNum()),
		txColor:    make(map[int64]int, g.TxNodeNum()),
		graph:      g,
	}
}

func (p *Partitioning) Set(n TxNode, color int) {
	if p.graph.IsBlockNode(n) {
		p.blockColor[p.graph.NodeIndex(n)] = color
	} else {
		p.txColor[p.graph.NodeIndex(n)] = color
	}
}

func (p *Partitioning) Get(n TxNode) int {
	if p.graph.IsBlockNode(n) {
		return p.blockColor[p.graph.NodeIndex(n)]
	} else {
		return p.txColor[p.graph.NodeIndex(n)]
	}
}
