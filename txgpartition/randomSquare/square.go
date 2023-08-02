package randomsquare

import (
	"math"
	"math/rand"
	"sort"
	"time"

	txp "github.com/tendermint/tendermint/txgpartition"
)

type SquareNode struct {
	id   int
	x, y float64
}

type RandomSquare struct {
	family                  []*SquareNode
	indegrees               []int
	outdegrees              []int
	fatherNodes             []map[int]txp.TxNode
	childNodes              []map[int]txp.TxNode
	visitMap                map[int]bool
	size                    int
	blockNodeFall           int
	blockNodeNum, txNodeNum int
	edgeNum                 int
}

var _ txp.TxGraph = (*RandomSquare)(nil)
var _ txp.TxNode = (*SquareNode)(nil)

// ================== CREATE NODE ======================================
func NewSquareNode(x, y float64) *SquareNode {
	return &SquareNode{
		x: x,
		y: y,
	}
}
func NewRandomSquareNode() *SquareNode {
	x, y := rand.Float64(), rand.Float64()
	return NewSquareNode(x, y)
}
func Distance(X, Y *SquareNode) float64 {
	dx, dy := X.x-Y.x, X.y-Y.y
	return math.Sqrt(dx*dx + dy*dy)
}
func ThresholdDistance(n int) float64 {
	return 0.45 * math.Sqrt(math.Log(float64(n))/float64(n))
}
func (n *SquareNode) SetID(id int) {
	n.id = id
}

// =================== CREATE GRAPH =====================================
func NewRandomSquare(n int) *RandomSquare {
	u := &RandomSquare{
		family:      make([]*SquareNode, n),
		indegrees:   make([]int, n),
		outdegrees:  make([]int, n),
		fatherNodes: make([]map[int]txp.TxNode, n),
		childNodes:  make([]map[int]txp.TxNode, n),
		visitMap:    map[int]bool{},
		size:        n,
	}
	for i := 0; i < n; i++ {
		u.fatherNodes[i] = make(map[int]txp.TxNode)
		u.childNodes[i] = make(map[int]txp.TxNode)
	}
	return u
}
func (rs *RandomSquare) RandomInit(blockRate float64) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < rs.size; i++ {
		rs.family[i] = NewRandomSquareNode()
	}
	sort.Slice(rs.family, func(i, j int) bool {
		return rs.family[i].x < rs.family[j].x || rs.family[i].x == rs.family[j].x && rs.family[i].y < rs.family[j].y
	})
	for i := 0; i < rs.size; i++ {
		rs.family[i].SetID(i)
	}
	TDis := ThresholdDistance(rs.size)
	for i := 0; i < rs.size; i++ {
		for j := i + 1; j < rs.size; j++ {
			if Distance(rs.family[i], rs.family[j]) < TDis {
				rs.outdegrees[j]++
				rs.indegrees[i]++
				rs.fatherNodes[j][i] = rs.family[i]
				rs.childNodes[i][j] = rs.family[j]
				rs.edgeNum++
			}
		}
	}
	zeroNum := len(rs.FindZeroOutdegree())
	rs.blockNodeNum = int(math.Floor(float64(zeroNum) * blockRate))
	rs.txNodeNum = rs.size - rs.blockNodeNum
	cnt := 0
	for i := 0; cnt < rs.blockNodeNum; i++ {
		if rs.outdegrees[i] == 0 {
			rs.blockNodeFall = i
			cnt++
		}
	}
}

// ================== NODE ==============================================
func (n *SquareNode) Less(other txp.TxNode) bool {
	e, ok := other.(*SquareNode)
	return ok && n.id < e.id
}
func (n *SquareNode) Equal(other txp.TxNode) bool {
	e, ok := other.(*SquareNode)
	return ok && n.id == e.id
}
func (n *SquareNode) ID() int {
	return n.id
}

// ==================== SQUARE ===========================================
func (rs *RandomSquare) IsBlockNode(n txp.TxNode) bool {
	return n.ID() <= rs.blockNodeFall
}
func (rs *RandomSquare) InDegree(n txp.TxNode) int {
	return rs.indegrees[n.ID()]
}
func (rs *RandomSquare) OutDegree(n txp.TxNode) int {
	return rs.outdegrees[n.ID()]
}
func (rs *RandomSquare) DecOutDegree(n txp.TxNode) {
	rs.outdegrees[n.ID()]--
}
func (rs *RandomSquare) Visited(n txp.TxNode) bool {
	return rs.visitMap[n.ID()]
}
func (rs *RandomSquare) Visit(n txp.TxNode) {
	rs.visitMap[n.ID()] = true
}
func (rs *RandomSquare) NodeIndex(n txp.TxNode) int {
	return n.ID()
}
func (rs *RandomSquare) BlockNodeNum() int {
	return rs.blockNodeNum
}
func (rs *RandomSquare) TxNodeNum() int {
	return rs.txNodeNum
}
func (rs *RandomSquare) FindZeroOutdegree() []txp.TxNode {
	out := make([]txp.TxNode, 0)
	for i, n := range rs.family {
		if rs.outdegrees[i] == 0 {
			out = append(out, n)
		}
	}
	return out
}
func (rs *RandomSquare) QueryFather(n txp.TxNode) map[int]txp.TxNode {
	return rs.fatherNodes[n.ID()]
}
func (rs *RandomSquare) QueryNodeChild(n txp.TxNode) map[int]txp.TxNode {
	return rs.childNodes[n.ID()]
}
