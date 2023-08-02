package txgpartition

import "fmt"

const (
	EdgeWeight = 1
)

func Init_Partitioning(g TxGraph, K int) (map[int]int, *ColorMap, [][]TxNode) {
	var (
		partitioning    = make(map[int]int)
		B               = g.BlockNodeNum()
		T               = g.TxNodeNum()
		colorMap        = CreateColorMap(B + K)
		BlockNodeQueue  = NewTxDequeue()
		TxNodeQueue     = NewTxDequeue()
		startColor      = B
		startBlockColor = 0
		cnt             = 0
		txMap           = make([][]TxNode, B+K)
	)

	// initialize 0 indegree nodes
	initNodes := g.FindZeroOutdegree()
	var tmpBlockSet, tmpTxSet []TxNode
	for _, n := range initNodes {
		if g.IsBlockNode(n) {
			tmpBlockSet = append(tmpBlockSet, n)
		} else {
			tmpTxSet = append(tmpTxSet, n)
		}
	}
	sortByOutDegreeLess := func(a, b TxNode) bool {
		return g.InDegree(a) < g.InDegree(b)
	}
	tmpBlockSet = TxSortBetterFirst(tmpBlockSet, sortByOutDegreeLess)
	tmpTxSet = TxSortBetterFirst(tmpTxSet, sortByOutDegreeLess)
	for _, n := range tmpBlockSet {
		BlockNodeQueue.PushBack(n)
	}
	BlockNodeQueue.PushBack(nil)

	//--------process----------------------------
	Lmax := T / K
	if T%K != 0 {
		Lmax++
	}
	for {
		if TxNodeQueue.Empty() {
			// all nodes have been processed
			if BlockNodeQueue.Empty() {
				break
			} else if b := BlockNodeQueue.PopFront(); b == nil {
				for _, n := range tmpTxSet {
					TxNodeQueue.PushBack(n)
				}
			} else if !g.Visited(b) {
				var tmp []TxNode
				for _, n := range g.QueryNodeChild(b) {
					g.DecOutDegree(n)
					if g.OutDegree(n) == 0 {
						tmp = append(tmp, n)
					}
				}
				// the more priceful, the latter in tmp list
				typeMap := make(map[int]int)
				childMap := make(map[int]int)
				for _, n := range tmp {
					seen := make(map[int]bool)
					nodeID := g.NodeIndex(n)
					for _, father := range g.QueryFather(n) {
						fatherColor := partitioning[g.NodeIndex(father)]
						if !seen[fatherColor] {
							seen[fatherColor] = true
							typeMap[nodeID]++
						}
					}
					childMap[nodeID] = g.InDegree(n)
				}
				compareBlockNodeMorePriceful := func(a, b TxNode) bool {
					if typeBeforeA, typeBeforeB := typeMap[g.NodeIndex(a)], typeMap[g.NodeIndex(b)]; typeBeforeA < typeBeforeB {
						return true
					} else if typeBeforeA > typeBeforeB {
						return false
					}
					if childrenNumA, childrenNumB := childMap[g.NodeIndex(a)], childMap[g.NodeIndex(b)]; childrenNumA > childrenNumB {
						return true
					} else if childrenNumA < childrenNumB {
						return false
					}
					return a.Less(b)
				}
				tmp = TxSortBetterFirst(tmp, compareBlockNodeMorePriceful)
				for _, n := range tmp {
					TxNodeQueue.PushFront(n)
				}
				g.Visit(b)
				partitioning[g.NodeIndex(b)] = startBlockColor
				txMap[startColor] = append(txMap[startBlockColor], b)
				startBlockColor++
			}
		} else if tx := TxNodeQueue.PopFront(); !g.Visited(tx) {
			// generate a color
			if cnt++; cnt > Lmax {
				cnt = 1
				startColor++
			}
			if startColor >= B+K {
				panic(g.NodeIndex(tx))
			}
			partitioning[g.NodeIndex(tx)] = startColor
			txMap[startColor] = append(txMap[startColor], tx)
			for _, n := range g.QueryFather(tx) {
				fatherColor := partitioning[g.NodeIndex(n)]
				colorMap.Add(fatherColor, startColor, EdgeWeight)
			}
			// add its children
			var tmp []TxNode
			for _, n := range g.QueryNodeChild(tx) {
				g.DecOutDegree(n)
				if g.OutDegree(n) == 0 {
					tmp = append(tmp, n)
				}
			}
			// the more priceful, the latter in tmp list
			costMap := make(map[int]int)
			typeMap := make(map[int]int)
			childMap := make(map[int]int)
			for _, n := range tmp {
				seen := make(map[int]bool)
				nodeID := g.NodeIndex(n)
				for _, father := range g.QueryFather(n) {
					fatherColor := partitioning[g.NodeIndex(father)]
					if !seen[fatherColor] {
						if colorMap.WillResultInEdgeAdd(fatherColor, startColor) {
							costMap[nodeID]++
						}
						seen[fatherColor] = true
						typeMap[nodeID]++
					}
				}
				childMap[nodeID] = g.InDegree(n)
			}
			compareNodeMorePriceful := func(a, b TxNode) bool {
				// 1  R(T,t)越小，优先级越高
				if costA, costB := costMap[g.NodeIndex(a)], costMap[g.NodeIndex(b)]; costA < costB {
					return true
				} else if costA > costB {
					return false
				}
				if typeBeforeA, typeBeforeB := typeMap[g.NodeIndex(a)], typeMap[g.NodeIndex(b)]; typeBeforeA > typeBeforeB {
					return true
				} else if typeBeforeA < typeBeforeB {
					return false
				}
				if childrenNumA, childrenNumB := childMap[g.NodeIndex(a)], childMap[g.NodeIndex(b)]; childrenNumA > childrenNumB {
					return true
				} else if childrenNumA < childrenNumB {
					return false
				}
				return a.Less(b)

			}
			tmp = TxSortBetterFirst(tmp, compareNodeMorePriceful)
			for _, n := range tmp {
				TxNodeQueue.PushFront(n)
			}
			g.Visit(tx)
		}
	}
	return partitioning, colorMap, txMap
}

func SimpleMove(g TxGraph, K int, alpha float64, partitioning map[int]int, colorMap *ColorMap, txMap [][]TxNode) (map[int]int, *ColorMap, [][]TxNode) {
	var (
		avaSize = float64(g.TxNodeNum()) / float64(K)
		Lmax    = (1.0 + alpha) * avaSize
		Lmin    = (1.0 - alpha) * avaSize
	)
	ColorSize := make([]int, len(txMap))
	for i := 0; i < len(txMap); i++ {
		ColorSize[i] = len(txMap[i])
	}
	moveTo := func(n TxNode, dst int, childs, fathers []int) {
		txID := g.NodeIndex(n)
		nodeColor := partitioning[txID]
		colorMap.ChangeTo(nodeColor, dst, childs, fathers)
		partitioning[txID] = dst
		ColorSize[nodeColor]--
		ColorSize[dst]++
	}

	for i := len(txMap) - 1; i >= 0; i-- {
		for _, n := range txMap[i] {
			if g.IsBlockNode(n) {
				continue
			}
			nodeColor := partitioning[g.NodeIndex(n)]
			if float64(ColorSize[nodeColor]) < Lmin {
				continue
			}
			fathers, childs := []int{}, []int{}
			moveToBigger := nodeColor+1 < len(txMap) && float64(ColorSize[nodeColor+1]) < Lmax
			moveToSmaller := nodeColor-1 >= g.BlockNodeNum() && float64(ColorSize[nodeColor-1]) < Lmax
			for _, father := range g.QueryFather(n) {
				if fatherColor := partitioning[g.NodeIndex(father)]; fatherColor == nodeColor {
					moveToSmaller = false
					break
				} else {
					fathers = append(fathers, fatherColor)
				}
			}
			for _, child := range g.QueryNodeChild(n) {
				if childColor := partitioning[g.NodeIndex(child)]; childColor == nodeColor {
					moveToBigger = false
					break
				} else {
					childs = append(childs, childColor)
				}
			}
			var scoreBigger, scoreSmaller = 0, 0
			if moveToBigger {
				scoreBigger = colorMap.Score(nodeColor, nodeColor+1, childs, fathers)
				moveToBigger = (scoreBigger > 0 || scoreBigger == 0 && ColorSize[nodeColor] > ColorSize[nodeColor+1])
			}
			if moveToSmaller {
				scoreSmaller = colorMap.Score(nodeColor, nodeColor-1, childs, fathers)
				moveToSmaller = scoreSmaller > 0 || scoreSmaller == 0 && ColorSize[nodeColor] > ColorSize[nodeColor-1]
			}
			fmt.Println(n.ID(), scoreBigger, scoreSmaller)
			var dst int
			if moveToBigger && moveToSmaller {
				if scoreSmaller > scoreBigger {
					dst = nodeColor - 1
				} else {
					dst = nodeColor + 1
				}
			} else if moveToBigger && scoreBigger > 0 {
				dst = nodeColor + 1
			} else if moveToSmaller && scoreSmaller > 0 {
				dst = nodeColor - 1
			} else {
				dst = -1
			}
			if dst > 0 {
				moveTo(n, dst, childs, fathers)
			}
		}
	}
	out := make([][]TxNode, len(txMap))
	for _, l1 := range txMap {
		for _, n := range l1 {
			color := partitioning[g.NodeIndex(n)]
			out[color] = append(out[color], n)
		}
	}
	return partitioning, colorMap, out
}
