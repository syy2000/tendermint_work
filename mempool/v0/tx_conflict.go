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
	seen := map[string]string{}
	for index, txObAndAttr := range memTx.tx.TxObAndAttr {
		op := memTx.tx.TxOp[index] // 读/写操作
		if op0, ok := seen[txObAndAttr]; ok {
			if op0 == "read" && op == "write" {
				seen[txObAndAttr] = op
			}
		} else {
			seen[txObAndAttr] = op
		}
	}
	for txObAndAttr, op := range seen {
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
				if len(conflictMapValue.RL) > 0 {
					for _, father := range conflictMapValue.RL {
						if father.isBlock {
							blockDepMap[father.ID()] = father
						} else {
							depMap[father.ID()] = father
						}
					}
				} else if len(conflictMapValue.WL) > 0 {
					for _, father := range conflictMapValue.WL {
						if father.isBlock {
							blockDepMap[father.ID()] = father
						} else {
							depMap[father.ID()] = father
						}
					}
				}
				conflictMapValue.RL = nil
				conflictMapValue.WL = []*mempoolTx{memTx}
			default:
				panic("while generating txMap -- procTxDEpendency : this should not happen! op must be read/write")
			}
		} else { // 事务依赖表没有此项，查找区块状态映射表
			if blockID, ok := mem.blockStatusMappingTable.Load(txObAndAttr); ok {
				//区块状态映射表中的区块号作为前序依赖事务
				blockTx, ok := mem.blockNodes[blockID]
				if !ok {
					blockTx = NewBlockMempoolTx(blockID)
					mem.blockNodes[blockID] = blockTx
					mem.blockNodeNum++
				}
				conflictMapValue := &txsConflictMapValue{
					RL: nil,
					WL: []*mempoolTx{blockTx},
				}
				mem.txsConflictMap[txObAndAttr] = conflictMapValue // 存入事务依赖表
				blockDepMap[blockTx.ID()] = blockTx
				// 事务依赖表和区块映射表均没有此项，新增至事务依赖表
			} else if op == "read" {
				conflictMapValue := &txsConflictMapValue{
					RL: []*mempoolTx{memTx},
					WL: nil,
				}
				mem.txsConflictMap[txObAndAttr] = conflictMapValue
			} else if op == "write" {
				conflictMapValue := &txsConflictMapValue{
					RL: []*mempoolTx{memTx},
					WL: nil,
				}
				mem.txsConflictMap[txObAndAttr] = conflictMapValue
			} else {
				panic("while generating txMap -- procTxDEpendency : this should not happen! op must be read/write")
			}
		}
	}
	//modified by syy
	//生成结点的邻接关系
	for _, father := range blockDepMap {
		GenEdge(father, memTx)
	}
	for _, father := range depMap {
		GenEdge(father, memTx)
	}
}
