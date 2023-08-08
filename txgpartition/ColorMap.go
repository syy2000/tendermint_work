package txgpartition

// ColorMap shuold only be used in txgpartition, so it will not check if i >= n or j >= n
// This approach is not safe, but quick
type ColorMap struct {
	adjMatrix [][]int
	size      int
}

func CreateColorMap(n int) *ColorMap {
	c := make([][]int, n)
	for i := 0; i < n; i++ {
		c[i] = make([]int, n)
	}
	return &ColorMap{
		adjMatrix: c,
		size:      n,
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

// only used in test;
// this function is meaningless.
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
