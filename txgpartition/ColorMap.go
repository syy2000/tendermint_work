package txgpartition

// ColorMap shuold only be used in txgpartition, so it will not check if i >= n or j >= n
// This approach is not safe, but quick
type ColorMap struct {
	adjMatrix     [][]int
	size          int
	numBlockColor int
	inDegrees     []int
	numTxBlocks   int
}

func CreateColorMap(n int, b int) *ColorMap {
	c := make([][]int, n)
	for i := 0; i < n; i++ {
		c[i] = make([]int, n)
	}
	return &ColorMap{
		adjMatrix:     c,
		size:          n,
		numBlockColor: b,
	}
}

func (c *ColorMap) Has(i, j int) bool {
	return c.adjMatrix[i][j] > 0
}

func (c *ColorMap) WillResultInEdgeAdd(i, j int) bool {
	return c.adjMatrix[i][j] == 0 && i != j
}

func (c *ColorMap) WillResultInEdgeDec(i, j, amount int) bool {
	return c.adjMatrix[i][j] == amount && i != j
}

// should only be called onece, when not all blocks have at least 1 txs, to delete some block parts
func (c *ColorMap) SetTxBlockNum(n int) {
	c.numTxBlocks, c.size = n, n+c.numBlockColor
}

func (c *ColorMap) Score(i, j int, inEdges, outEdges []int) int {
	var ScoreDec, ScoreAdd int
	inEdgeMap := make(map[int]int)
	outEdgeMap := make(map[int]int)
	for _, inEdge := range inEdges {
		inEdgeMap[inEdge] = inEdgeMap[inEdge] + 1
	}
	for _, outEdge := range outEdges {
		outEdgeMap[outEdge] = outEdgeMap[outEdge] + 1
	}
	for inEdge, edgeAmount := range inEdges {
		if c.WillResultInEdgeAdd(j, inEdge) {
			ScoreAdd++
		}
		if c.WillResultInEdgeDec(i, inEdge, edgeAmount) {
			ScoreDec++
		}
	}
	for outEdge, edgeAmount := range outEdges {
		if c.WillResultInEdgeAdd(outEdge, j) {
			ScoreAdd++
		}
		if c.WillResultInEdgeDec(outEdge, i, edgeAmount) {
			ScoreDec++
		}
	}
	return ScoreDec - ScoreAdd
}

func (c *ColorMap) Add(i, j, amount int) {
	if i != j {
		c.adjMatrix[i][j] += amount
	}
}

func (c *ColorMap) Sub(i, j, amount int) {
	if i != j {
		c.adjMatrix[i][j] -= amount
	}
}

// move node with inEdges and outEdges from i to j
func (c *ColorMap) ChangeTo(i, j int, inEdges, outEdges []int) {
	for _, inEdge := range inEdges {
		c.Sub(i, inEdge, EdgeWeight)
		c.Add(j, inEdge, EdgeWeight)
	}
	for _, outEdge := range outEdges {
		c.Sub(outEdge, i, EdgeWeight)
		c.Add(outEdge, j, EdgeWeight)
	}
}

// calculate colorMap's indegree; must be done whenever finishing partitioning
func (c *ColorMap) CalIndegrees() {
	c.ClearInDegree()
	var u int
	for i := c.numBlockColor; i < c.size; i++ {
		u = 0
		for j := c.numBlockColor; j < c.size; j++ {
			u += c.adjMatrix[j][i]
		}
		c.SetIndegree(i, u)
	}
}

func (c *ColorMap) ReapBlocks(n int) (int, []int) {
	var (
		truen int   = 0
		out   []int = make([]int, 0, n)
	)
	for i := c.numBlockColor; i < c.size && truen < n; i++ {
		if c.Indegree(i) == 0 {
			truen++
			c.SetIndegree(i, -1)
			out = append(out, i)
		}
	}
	for _, i := range out {
		for j := i + 1; j < c.size; j++ {
			c.DecInDegree(j, c.adjMatrix[i][j])
		}
	}
	return truen, out
}

// ============================ tools =======================================

func (c *ColorMap) Indegree(i int) int {
	if i < c.numBlockColor {
		return 0
	}
	return c.inDegrees[i-c.numBlockColor]
}
func (c *ColorMap) DecInDegree(i int, amount int) {
	if i >= c.numBlockColor {
		c.inDegrees[i-c.numBlockColor] -= amount
	}
}
func (c *ColorMap) ClearInDegree() {
	c.inDegrees = make([]int, c.size-c.numBlockColor)
}
func (c *ColorMap) SetIndegree(i int, amount int) {
	if i >= c.numBlockColor {
		c.inDegrees[i-c.numBlockColor] = amount
	}
}

// =========================== test purpose =============================
func CalculatePartitioningQualityByColorMap(c *ColorMap) int {
	u := 0
	for i := 0; i < c.size; i++ {
		for j := 0; j < c.size; j++ {
			if c.Has(j, i) && i != j {
				u++
			}
		}
	}
	return u
}
func CalculatePartitioningQualityByInnerPartitioningEdge(c *ColorMap) int {
	var u int
	for i := 0; i < c.size; i++ {
		for j := 0; j < c.size; j++ {
			if i == j {
				continue
			}
			u += c.adjMatrix[i][j]
		}
	}
	return u
}
