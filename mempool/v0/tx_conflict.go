package v0

import (
	//"fmt"
	"runtime"
	"sync"
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
				// 清空读集，以后的读/写都和此写冲突
				conflictMapValue.RL = nil
				conflictMapValue.WL = []*mempoolTx{memTx}
			default:
				panic("while generating txMap -- procTxDEpendency : this should not happen! op must be read/write")
			}
		} else { // 事务依赖表没有此项，查找区块状态映射表
			// if blockID, ok := mem.blockStatusMappingTable.Load(txObAndAttr); ok {
			// 	//区块状态映射表中的区块号作为前序依赖事务
			// 	blockTx, ok := mem.blockNodes[blockID]
			// 	if !ok {
			// 		blockTx = NewBlockMempoolTx(blockID)
			// 		mem.blockNodes[blockID] = blockTx
			// 		mem.blockNodeNum++
			// 	}
			// 	conflictMapValue := &txsConflictMapValue{
			// 		RL: nil,
			// 		WL: []*mempoolTx{blockTx},
			// 	}
			// 	mem.txsConflictMap[txObAndAttr] = conflictMapValue // 存入事务依赖表
			// 	blockDepMap[blockTx.ID()] = blockTx
			// 	// 事务依赖表和区块映射表均没有此项，新增至事务依赖表
			// } else
			if op == "read" {
				conflictMapValue := &txsConflictMapValue{
					RL: []*mempoolTx{memTx},
					WL: nil,
				}
				mem.txsConflictMap[txObAndAttr] = conflictMapValue
				//fmt.Println("first read")
			} else if op == "write" {
				conflictMapValue := &txsConflictMapValue{
					RL: nil,
					WL: []*mempoolTx{memTx},
				}
				mem.txsConflictMap[txObAndAttr] = conflictMapValue
				//fmt.Println("first write")
			} else {
				panic("while generating txMap -- procTxDEpendency : this should not happen! op must be read/write")
			}
		}
	}

	//modified by syy
	//生成结点的邻接关系
	for _, father := range blockDepMap {
		GenEdge(father, memTx)
		//fmt.Println("hhh")
	}
	for _, father := range depMap {
		GenEdge(father, memTx)
	}
}

func GenEdge(father, child *mempoolTx) {
	father.childTxs = append(father.childTxs, child)
	child.parentTxs = append(child.parentTxs, father)
}

// diploma design
func (mem *CListMempool) ZeroOutDegreeMempoolTx(visit [40000]bool) []*mempoolTx {
	out := make([]*mempoolTx, 0)
	for _, tx := range mem.workspace {
		if tx.outDegree == 0 && !visit[tx.ID()] {
			out = append(out, tx)
		}
	}
	//fmt.Println(len(out))
	return out
}
func (mem *CListMempool) ExecuteSequentially(accountMap sync.Map) float64 {
	start := time.Now()
	//record_time := start
	for _, tx := range mem.workspace {
		for index, op := range tx.tx.TxOp {
			s := tx.tx.TxObAndAttr
			if op == "read" {
				_, _ = accountMap.Load(s[index])
				//fmt.Printf("%d", tmp)
			} else if op == "write" {
				cur, isFind := accountMap.Load(s[index])
				if isFind {
					accountMap.Store(s[index], cur.(int)+100)
				}
			}
		}
		//time_used_perTx := time.Since(record_time)
		//record_time = time.Now()
		//fmt.Printf("%.2f\n", float64(time_used_perTx)/float64(time.Millisecond))
	}
	time_used := time.Since(start)
	//fmt.Printf("%.2f\n", float64(time_used)/float64(time.Millisecond))
	return float64(time_used) / float64(time.Millisecond)
}
func doTask(tx *mempoolTx, accountMap sync.Map) {
	//wg *sync.WaitGroup,
	for index, op := range tx.tx.TxOp {
		s := tx.tx.TxObAndAttr
		if op == "read" {
			_, _ = accountMap.Load(s[index])
			//fmt.Printf("%d", tmp)
		} else if op == "write" {
			cur, isFind := accountMap.Load(s[index])
			if isFind {
				accountMap.Store(s[index], cur.(int)+100)
			}
		}
	}
	//(*wg).Done()
}
func doTasks(x int, out []*mempoolTx, accountMap sync.Map) {
	runtime.GOMAXPROCS(x)
	//var wg sync.WaitGroup
	//start := time.Now().Unix()
	for index, _ := range out {
		//wg.Add(1)
		//go doTask(&wg, out[index], accountMap)
		go doTask(out[index], accountMap)
	}
	//wg.Wait()
	//fmt.Println("cpu", x)
	//fmt.Println((time.Now().UnixNano()-start)/1e9)
}
func (mem *CListMempool) ExecuteConcurrently(accountMap sync.Map) float64 {
	//start := time.Now()
	txNum := len(mem.workspace)
	var start time.Time
	var time_used float64
	var out []*mempoolTx
	var i int
	var visit [40000]bool
	//拓扑+并发
	for txNum > 0 {
		//fmt.Println("The", i, "time")
		i += 1
		out = mem.ZeroOutDegreeMempoolTx(visit)
		//fmt.Println(len(out))
		start = time.Now()
		doTasks(8, out, accountMap)
		time_used += float64(time.Since(start))
		//子节点入度-1
		for _, tx := range out {
			childTxs := tx.childTxs
			for _, child := range childTxs {
				mem.workspace[child.ID()].outDegree -= 1
			}
			visit[tx.ID()] = true
		}
		txNum -= len(out)
		//start = time.Now()
	}
	return float64(time_used) / float64(time.Millisecond)
}
