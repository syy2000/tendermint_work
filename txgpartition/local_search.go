package txgpartition

const (
	EdgeWeight = 1
)

/*
partitioning : map[int64]int : 事务id -> 事务颜色
colorMap : *ColorMap : 事务颜色的邻接矩阵
txMap : TransactionMap : 事务颜色(事务集编号) -> 属于该颜色的事务集合
*/

func Init_Partitioning(g TxGraph, K int, _ float64) (*Partitioning, *ColorMap, *TransactionMap) {
	var (
		partitioning    = NewPartitioning(g)
		B               = g.BlockNodeNum()
		T               = g.TxNodeNum()
		colorMap        = CreateColorMap(K, B)
		BlockNodeQueue  = NewTxDequeue()
		TxNodeQueue     = NewTxDequeue()
		startColor      = B
		startBlockColor = 0
		cnt             = 0
		txMap           = NewTransactionMap(B, K, T/K+1)
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
				// update color
				if cnt+1 > Lmax {
					cnt = 0
					startColor++
				}
				// choose children
				bschild := g.QueryNodeChild(b)
				tmp := make([]TxNode, 0, len(bschild))
				for _, n := range bschild {
					g.DecOutDegree(n)
					if g.OutDegree(n) == 0 {
						tmp = append(tmp, n)
					}
				}
				// the more priceful, the latter in tmp list
				costMap := make(map[int64]int)
				typeMap := make(map[int64]int)
				childMap := make(map[int64]int)
				for _, n := range tmp {
					seen := make(map[int]bool)
					nodeID := g.NodeIndex(n)
					for _, father := range g.QueryFather(n) {
						fatherColor := partitioning.Get(father)
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
				compareBlockNodeMorePriceful := func(a, b TxNode) bool {
					if costA, costB := costMap[g.NodeIndex(a)], costMap[g.NodeIndex(b)]; costA < costB {
						return true
					} else if costA > costB {
						return false
					}
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
				partitioning.Set(b, startBlockColor)
				txMap.Append(startBlockColor, b)
				startBlockColor++
			}
		} else if tx := TxNodeQueue.PopFront(); !g.Visited(tx) {
			// generate a color
			if startColor >= B+K {
				panic(g.NodeIndex(tx))
			}
			partitioning.Set(tx, startColor)
			txMap.Append(startColor, tx)
			for _, n := range g.QueryFather(tx) {
				fatherColor := partitioning.Get(n)
				colorMap.Add(fatherColor, startColor, EdgeWeight)
			}
			cnt++
			if cnt+1 > Lmax {
				cnt = 0
				startColor++
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
			costMap := make(map[int64]int)
			typeMap := make(map[int64]int)
			childMap := make(map[int64]int)
			for _, n := range tmp {
				seen := make(map[int]bool)
				nodeID := g.NodeIndex(n)
				for _, father := range g.QueryFather(n) {
					fatherColor := partitioning.Get(father)
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
	// calculate colorMap's indegree
	colorMap.CalIndegrees()
	return partitioning, colorMap, txMap
}

func SimpleMove(g TxGraph, K int, alpha float64, partitioning *Partitioning, colorMap *ColorMap, txMap *TransactionMap) (*Partitioning, *ColorMap, *TransactionMap) {
	var (
		avaSize = float64(g.TxNodeNum()) / float64(K)
		Lmax    = (1.0 + alpha) * avaSize
		Lmin    = (1.0 - alpha) * avaSize
	)
	ColorSize := make([]int, txMap.Size())
	for i := 0; i < txMap.Size(); i++ {
		ColorSize[i] = len(txMap.Get(i))
	}
	moveTo := func(n TxNode, dst int, childs, fathers []int) {
		nodeColor := partitioning.Get(n)
		colorMap.ChangeTo(nodeColor, dst, childs, fathers)
		partitioning.Set(n, dst)
		ColorSize[nodeColor]--
		ColorSize[dst]++
	}

	for i := txMap.Size() - 1; i >= 0; i-- {
		for _, n := range txMap.Get(i) {
			if g.IsBlockNode(n) {
				continue
			}
			nodeColor := partitioning.Get(n)
			if float64(ColorSize[nodeColor]) < Lmin {
				continue
			}
			FatherSet, ChildSet := g.QueryFather(n), g.QueryNodeChild(n)
			fathers, childs := make([]int, len(FatherSet)), make([]int, len(ChildSet))
			moveToBigger := nodeColor+1 < txMap.Size() && float64(ColorSize[nodeColor+1]) < Lmax-1.0
			moveToSmaller := nodeColor-1 >= g.BlockNodeNum() && float64(ColorSize[nodeColor-1]) < Lmax-1.0
			j := 0
			var biggerCut, smallerCut, selfCut int
			for _, father := range FatherSet {
				fatherColor := partitioning.Get(father)
				if fatherColor == nodeColor {
					moveToSmaller = false
					selfCut++
				}
				if fatherColor == nodeColor-1 {
					smallerCut++
				}
				if fatherColor == nodeColor+1 {
					biggerCut++
				}
				fathers[j] = fatherColor
				j++
			}
			j = 0
			for _, child := range ChildSet {
				childColor := partitioning.Get(child)
				if childColor == nodeColor {
					moveToBigger = false
					selfCut++
				}
				if childColor == nodeColor-1 {
					smallerCut++
				}
				if childColor == nodeColor+1 {
					biggerCut++
				}
				childs[j] = childColor
				j++
			}
			var scoreBigger, scoreSmaller = 0, 0
			dbiggerCut := biggerCut - selfCut
			dsmallerCut := smallerCut - selfCut
			if moveToBigger {
				scoreBigger = colorMap.Score(nodeColor, nodeColor+1, childs, fathers)
				moveToBigger = (scoreBigger > 0 ||
					scoreBigger == 0 && dbiggerCut > 0 ||
					scoreBigger == 0 && dbiggerCut == 0 && ColorSize[nodeColor] > ColorSize[nodeColor+1])
			}
			if moveToSmaller {
				scoreSmaller = colorMap.Score(nodeColor, nodeColor-1, childs, fathers)
				moveToSmaller = scoreSmaller > 0 ||
					scoreSmaller == 0 && dsmallerCut > 0 ||
					scoreSmaller == 0 && dsmallerCut == 0 && ColorSize[nodeColor] > ColorSize[nodeColor-1]
			}
			var dst int
			compareScoreSmallerBetter := moveToBigger && moveToSmaller &&
				(scoreSmaller > scoreBigger ||
					scoreSmaller == scoreBigger && dsmallerCut > dbiggerCut ||
					scoreSmaller == scoreBigger && dsmallerCut == dbiggerCut && ColorSize[nodeColor-1] < ColorSize[nodeColor+1])
			if moveToBigger && moveToSmaller {
				if compareScoreSmallerBetter {
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
	out := NewTransactionMap(txMap.blockNum, txMap.partitionNum, txMap.avasize)
	for l1 := 0; l1 < txMap.Size(); l1++ {
		for _, n := range txMap.Get(l1) {
			color := partitioning.Get(n)
			out.Append(color, n)
		}
	}
	// calculate colorMap's indegree
	colorMap.CalIndegrees()
	return partitioning, colorMap, out
}

func AdvancedMove(g TxGraph, K int, alpha float64, partitioning *Partitioning, colorMap *ColorMap, txMap *TransactionMap) (*Partitioning, *ColorMap, *TransactionMap) {
	var (
		avaSize = float64(g.TxNodeNum()) / float64(K)
		Lmax    = (1.0 + alpha) * avaSize
		Lmin    = (1.0 - alpha) * avaSize
	)
	ColorSize := make([]int, txMap.Size())
	for i := 0; i < txMap.Size(); i++ {
		ColorSize[i] = len(txMap.Get(i))
	}
	moveTo := func(n TxNode, dst int, childs, fathers []int) {
		nodeColor := partitioning.Get(n)
		colorMap.ChangeTo(nodeColor, dst, childs, fathers)
		partitioning.Set(n, dst)
		ColorSize[nodeColor]--
		ColorSize[dst]++
	}
	calculateRangeAndColor := func(n TxNode) (int, int, []int, []int) {
		A, B := g.BlockNodeNum(), colorMap.size-1
		fathers, childs := g.QueryFather(n), g.QueryNodeChild(n)
		fatherColors, childColors := make([]int, len(fathers)), make([]int, len(childs))
		i := 0
		for _, father := range fathers {
			fathercolor := partitioning.Get(father)
			if fathercolor > A {
				A = fathercolor
			}
			fatherColors[i] = fathercolor
			i++
		}
		i = 0
		for _, child := range childs {
			childcolor := partitioning.Get(child)
			if childcolor < B {
				B = childcolor
			}
			childColors[i] = childcolor
			i++
		}
		return A, B, fatherColors, childColors
	}
	for i := txMap.Size() - 1; i >= 0; i-- {
		for _, n := range txMap.Get(i) {
			if g.IsBlockNode(n) {
				continue
			}
			nodeColor := partitioning.Get(n)
			colorSizeNow := ColorSize[nodeColor]
			if float64(colorSizeNow)-1.0 < Lmin {
				continue
			}
			A, B, fatherColor, childColor := calculateRangeAndColor(n)
			maxp, maxscore, cutDec, delta := -1, 0, 0, 0
			for j := A; j < B; j++ {
				if j == nodeColor || float64(ColorSize[j])+1.0 > Lmax {
					continue
				}
				s := colorMap.Score(nodeColor, j, childColor, fatherColor)
				// calculate cut
				cd := 0
				for _, color := range fatherColor {
					if color == j {
						cd++
					}
					if color == nodeColor {
						cd--
					}
				}
				for _, color := range childColor {
					if color == j {
						cd++
					}
					if color == nodeColor {
						cd--
					}
				}
				// choose best
				betterChoiceWitness := s > maxscore ||
					s == maxscore && cd > cutDec ||
					s == maxscore && cd == cutDec && colorSizeNow-ColorSize[j] > delta
				if betterChoiceWitness {
					maxp = j
					maxscore = s
					cutDec = cd
					delta = colorSizeNow - ColorSize[j]

				}
			}
			if maxp != -1 && maxscore >= 0 {
				moveTo(n, maxp, childColor, fatherColor)
			}
		}
	}
	out := NewTransactionMap(txMap.blockNum, txMap.partitionNum, txMap.avasize)
	for l1 := 0; l1 < txMap.Size(); l1++ {
		for _, n := range txMap.Get(l1) {
			color := partitioning.Get(n)
			out.Append(color, n)
		}
	}
	// calculate colorMap's indegree
	colorMap.CalIndegrees()
	return partitioning, colorMap, out
}
