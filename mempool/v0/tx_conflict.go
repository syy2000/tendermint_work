package v0

import (
	"fmt"
	"time"
)

var tmused time.Duration

// modified by syy
type txsConflictMapValue struct {
	RL []*mempoolTx
	WL []*mempoolTx
}

func (mem *CListMempool) ProcWorkspaceDependency() {
	for i, tx := range mem.workspace {
		tx.tx.TxId = int64(i)
		mem.procTxDependency(tx)
	}
	for _, tx := range mem.blockNodes {
		tx.inDegree = len(tx.childTxs)
	}
	for _, tx := range mem.workspace {
		tx.inDegree = len(tx.childTxs)
		tx.outDegree = len(tx.parentTxs)
	}
}

// donghao =========================================================

func (mem *CListMempool) procTxDependency(memTx *mempoolTx) {
	//fistly modified by syy
	//查找事务依赖表，找到“对象id+属性”的前序依赖
	blockDepMap, depMap := map[int64]*mempoolTx{}, map[int64]*mempoolTx{}
	seen := map[string]string{}
	start := time.Now()
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
	tmused += time.Since(start)
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

func GenEdge(father, child *mempoolTx) {
	father.childTxs = append(father.childTxs, child)
	child.parentTxs = append(child.parentTxs, father)
}

//diploma design
func (mem *CListMempool) ZeroOutDegreeMempoolTx() []*mempoolTx {
	out := make([]*mempoolTx, 0)
	for _, tx := range mem.workspace {
		if len(tx.childTxs) == 0 {
			out = append(out, tx)
		}
	}
	return out
}
func (mem *CListMempool) ExecuteSequentially(accountMap map[string]int64) float64 {
	start := time.Now()
	//record_time := start
	for _,tx := range mem.workspace { 
		for index,op := range tx.tx.TxOp {
			s := tx.tx.TxObAndAttr
			if op == "read" {
				_ = accountMap[s[index]]
				//fmt.Printf("%d", tmp)
			} else if op == "write" {
				accountMap[s[index]] += 100
			}
		}
		//time_used_perTx := time.Since(record_time)
		//record_time = time.Now()
		//fmt.Printf("%.2f\n", float64(time_used_perTx)/float64(time.Millisecond))
	}
	time_used := time.Since(start)
	fmt.Printf("%.2f\n", float64(time_used)/float64(time.Millisecond))
	return float64(time_used)/float64(time.Millisecond)
}

func (mem *CListMempool) ExecuteConcurrently(accountMap map[string]int64) float64 {
	start := time.Now()
	out := mem.ZeroOutDegreeMempoolTx()
	 
}