package txgpartition

func TxSortBetterFirst(txNodeList []TxNode, Less func(TxNode, TxNode) bool) []TxNode {
	txSortBetterFirst(txNodeList, Less, 0, len(txNodeList))
	return txNodeList
}

func txSortBetterFirst(txNodeList []TxNode, Less func(TxNode, TxNode) bool, start, end int) {
	if end-start <= 1 {
		return
	}
	mid := (start + end) / 2
	txSortBetterFirst(txNodeList, Less, start, mid)
	txSortBetterFirst(txNodeList, Less, mid, end)
	var (
		i   = start
		j   = mid
		k   = 0
		tmp = make([]TxNode, end-start)
	)
	for i < mid && j < end {
		if Less(txNodeList[i], txNodeList[j]) {
			tmp[k] = txNodeList[j]
			j++
		} else {
			tmp[k] = txNodeList[i]
			i++
		}
		k++
	}
	for i < mid {
		tmp[k] = txNodeList[i]
		i++
		k++
	}
	for j < end {
		tmp[k] = txNodeList[j]
		j++
		k++
	}
	for i := 0; i < len(tmp); i++ {
		txNodeList[i+start] = tmp[i]
	}
}
