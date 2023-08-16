package v0

// modified by syy
type txsConflictMapValue struct {
	RL []*mempoolTx
	WL []*mempoolTx
}

func GenEdge(father, child *mempoolTx) {
	father.childTxs = append(father.childTxs, child)
	father.inDegree++
	child.parentTxs = append(child.parentTxs, father)
	child.outDegree++
}

// donghao =========================================================

func (mem *CListMempool) procTxDependency(memTx *mempoolTx) {
	//fistly modified by syy
	//查找事务依赖表，找到“对象id+属性”的前序依赖
	blockDepMap, depMap := map[int64]*mempoolTx{}, map[int64]*mempoolTx{}
	for index, txObAndAttr := range memTx.tx.TxObAndAttr {
		op := memTx.tx.TxOp[index] // 读/写操作

		if conflictMapValue, ok := mem.txsConflictMap[txObAndAttr]; ok {
			switch op {
			case "read":
				if len(conflictMapValue.WL) > 0 {
					for _, father := range conflictMapValue.WL {
						if father.isBlock {
							blockDepMap[father.ID()] = father
						} else {
							depMap[father.ID()] = father
						}
					}
				}
				conflictMapValue.RL = append(conflictMapValue.RL, memTx)
			case "write":
			default:
				panic("while generating txMap -- procTxDEpendency : this should not happen! op must be read/write")
			}
		} else { // 事务依赖表没有此项，查找区块状态映射表
			if v, ok := mem.blockStatusMappingTable.Load(txObAndAttr); ok {
				//区块状态映射表中的区块号作为前序依赖事务
				blockId := v
				prevTx := &mempoolTx{
					isBlock: true,
				}
				prevTx.tx.SetTxId(blockId)
				conflictMapValue := &txsConflictMapValue{
					attrValue: "",
					curTx:     []*mempoolTx{memTx},
					prevTx:    []*mempoolTx{prevTx},
					operation: op,
				}
				mem.txsConflictMap.Store(txObAndAttr, conflictMapValue) // 存入事务依赖表
			} else { // 事务依赖表和区块映射表均没有此项，新增至事务依赖表
				conflictMapValue := &txsConflictMapValue{
					attrValue: "",
					curTx:     []*mempoolTx{memTx},
					// prevTx:    []*mempoolTx{prevTx},
					operation: op,
				}
				mem.txsConflictMap.Store(txObAndAttr, conflictMapValue)
			}
		}
	}
	//modified by syy
	//生成结点的邻接关系
	for _, txObAndAttr := range memTx.tx.TxObAndAttr {
		if v, ok := mem.txsConflictMap.Load(txObAndAttr); ok {
			//arr := strings.Split(v.(string), " ")
			conflictMapValue := v.(*txsConflictMapValue)
			//txArr := conflictMapValue.curTx
			prevTxs := conflictMapValue.prevTx //当前事务在此对象+属性上的前序依赖事务
			for _, prevTx := range prevTxs {
				// if !contains(memTx.conflictTxs, prevTx){
				// 	memTx.conflictTxs = append(memTx.conflictTxs, prevTx)
				// }
				if _, ok := memTx.parentTxs[prevTx.ID()]; !ok {
					memTx.parentTxs[prevTx.ID()] = prevTx
					memTx.outDegree += 1
				}
				//根据事务id找到对应的mempoolTx
				if _, ok := prevTx.childTxs[memTx.ID()]; !ok {
					prevTx.childTxs[memTx.ID()] = memTx
					prevTx.inDegree += 1
				}
			}
		}
	}
}
