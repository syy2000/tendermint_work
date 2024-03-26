package v0

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/tendermint/tendermint/txgpartition/statustable"
	"github.com/tendermint/tendermint/types"
	"github.com/xuri/excelize/v2"
)

const (
	keyNumPerTx = 4
	accountNum  = 10000
	hot_rate    = 0.8
	preBlockNum = 80
	testTimes   = 1
)

var (
	step = 4
)

var (
	xlsxPath        = "./mempool_testout.xlsx"
	writeXlsxFunc   func(int, int, string) error
	closeXlsxFunc   func()
	newSheetHandler func(string)
	xlsxSheetName   = "Test"
	accountMap      sync.Map
)

func TestMain(t *testing.T) {
	writeXlsxFunc = CreateOrOpen()
	defer closeXlsxFunc()
	writeXlsxFunc(2, 1, "time_used")
	writeXlsxFunc(3, 1, "edges")
	writeXlsxFunc(4, 1, "zero_outdegree")
	writeXlsxFunc(5, 1, "max_deps")
	writeXlsxFunc(6, 1, "mid_deps")
	writeXlsxFunc(7, 1, "countComponent")
	writeXlsxFunc(8, 1, "weight")
	//diploma design
	//添加账户余额表，模拟执行，read为直接访问,write为修改值
	//accountMap = make(map[string]int64, accountNum)
	for i := 0; i < accountNum; i++ {
		accountMap.Store(strconv.Itoa(i), 10000)
	}
	for i := 1; i <= 1; i++ {
		total := step * i * 100
		fmt.Printf("========= Node Num : %d ==========\n", total)
		writeXlsxFunc(1, i+1, fmt.Sprint(total))
		var (
			time_used                                              time.Duration
			edges, zero_outdegree, max_deps, mid_deps, totalWeight int
			Sequential_total_time, Concurrent_total_time           float64
			//Concurrent_total_time float64
			//componentMap map[int64][]int64
			//weightMap    map[int64]int64
			count int64
		)
		for i := 0; i < testTimes; i++ {
			mem := CListMempool{}
			mem.blockNodeNum = 0
			mem.blockNodes = map[int64]*mempoolTx{}
			mem.blockIDMap = make(map[int]int64)
			mem.txsConflictMap = make(map[string]*txsConflictMapValue)

			mem.workspace = GenReadWrites(total)
			mem.txNodeNum = total
			mem.blockStatusMappingTable = *GenBlockTable()

			fmt.Printf("start round %d\n", i)
			start := time.Now()
			mem.ProcWorkspaceDependency()
			mem.CalculateWeight() // 给mempool的workspace中所有mempoolTx加权
			time_used += time.Since(start)
			edges += mem.edgeNum()
			zero_outdegree = len(mem.FindZeroOutdegree()) - preBlockNum
			max_deps = mem.maxDep()
			mid_deps = mem.midDep()
			// for _, tx := range mem.workspace {
			// 	fmt.Println(tx.weight)
			// }
			n := 5
			count = mem.CountComponent() //数数mempool里面几个连通分量
			fmt.Println(count)
			totalWeight := mem.countWeight()
			componentMap, weightMap, locate := mem.DivideGraph(totalWeight, int64(n))
			mem.SeparateGraph(locate) // 按照每个连通分量的划分情况更新每个结点的childTxs和parentTxs
			//fmt.Println(totalWeight)
			for i := 0; i < 5; i++ {
				fmt.Printf("length of componentMap %d is %d\n", i, len(componentMap[int64(i)]))
				fmt.Printf("weight is %d\n", weightMap[int64(i)])
			}
			count = mem.CountComponent()
			fmt.Println(count)
			//fmt.Println(count)
			//mem.ExecuteConcurrently(accountMap)
			//Sequential_total_time += mem.ExecuteSequentially(accountMap)
			//Concurrent_total_time += mem.ExecuteConcurrently(accountMap)
			//Sequential_total_time += mem.ExecuteSequentially(accountMap)
			//_, outNodeSets := mem.BalanceReapBlocks(componentMap, weightMap, 20)
			// for _, txs := range outNodeSets {
			// 	// for _, tx := range txs {
			// 	// 	fmt.Printf("%s ", tx)
			// 	// }
			// 	fmt.Println(len(txs))
			// }
			// fmt.Println(len(outNodeSets))
		}
		time_used /= time.Duration(testTimes)
		writeXlsxFunc(2, i+1, fmt.Sprintf("%.2f", float64(time_used)/float64(time.Millisecond)))
		writeXlsxFunc(3, i+1, fmt.Sprint(edges/testTimes))
		writeXlsxFunc(4, i+1, fmt.Sprint(zero_outdegree))
		writeXlsxFunc(5, i+1, fmt.Sprint(max_deps))
		writeXlsxFunc(6, i+1, fmt.Sprint(mid_deps))
		writeXlsxFunc(7, i+1, fmt.Sprint(count))
		writeXlsxFunc(8, i+1, fmt.Sprint(totalWeight/testTimes))
		fmt.Println("Sequential Execute average time is", Sequential_total_time/testTimes)
		fmt.Println("Concurrent Execute average time is", Concurrent_total_time/testTimes)
		//var cur int64
		// for cur = 1; cur <= count; cur++ {
		// 	fmt.Println("component:", len(componentMap[cur]), " weight:", weightMap[cur])
		// }
		//fmt.Println(count)
	}
}

func GenReadWrites(n int) []*mempoolTx {
	//s := "hello"
	out := make([]*mempoolTx, n)
	for i := 0; i < n; i++ {
		var weight int64
		memTx := types.MemTx{
			TxId:        int64(i),
			TxOp:        []string{},
			TxObAndAttr: []string{},
			// Origin Tx
			OriginTx: types.Tx{OriginTx: []byte{'h', 'e', 'l', 'l', 'o'}},
		}
		keys := GenRandomKey(i, n)
		for j := 0; j < len(keys); j++ {
			randNum := rand.Intn(2)
			if randNum == 0 {
				memTx.TxOp = append(memTx.TxOp, "read")
				//weight += int64(rand.Intn(100))
			} else {
				memTx.TxOp = append(memTx.TxOp, "write")
				//weight += int64(rand.Intn(100) + (n-i)/10)
			}
			memTx.TxObAndAttr = append(memTx.TxObAndAttr, keys[j])
		}
		out[i] = NewMempoolTx(&memTx)
		out[i].weight = int64(weight)
	}
	return out
}
func GenBlockTable() *statustable.BlockStatusMappingTable {
	out := statustable.NewBlockStatusMappingTable(statustable.UseSimpleMap, nil)
	for i := 0; i < int(math.Ceil(float64(accountNum)*hot_rate)); i++ {
		out.Store(fmt.Sprint(i), int64(i%preBlockNum))
	}
	return out
}
func TestRandomKey(t *testing.T) {
	keys := GenRandomKey(100, 40000)
	for i := 0; i < len(keys); i++ {
		fmt.Printf("%s  ", keys[i])
	}
}
func GenRandomKey(i int, n int) []string {
	outInt := []int{}
	for len(outInt) < keyNumPerTx {
		x := rand.Intn(n)
		flg := false
		for _, y := range outInt {
			if x == y {
				flg = true // 每个mempoolTx最多访问特定txObAndAttr一次
			}
		}
		if flg {
			continue
		} else {
			outInt = append(outInt, x)
		}
	}

	out := make([]string, len(outInt))
	for i, x := range outInt {
		out[i] = strconv.Itoa(x)
	}
	return out
}

func CreateOrOpen() func(int, int, string) error {
	if exists, err := statustable.PathExists(xlsxPath); err != nil {
		panic(err)
	} else if exists {
		f, err := excelize.OpenFile(xlsxPath)
		if err != nil {
			panic(err)
		}
		closeXlsxFunc = func() {
			if err := f.SaveAs(xlsxPath); err != nil {
				panic(err)
			}
			f.Close()
		}
		newSheetHandler = func(a string) {
			if _, err := f.NewSheet(a); err != nil {
				panic(err)
			}
		}
		return func(line, col int, a string) error {
			cellID := fmt.Sprintf("%c%d", int8(col-1)+'A', line)
			return f.SetCellValue(xlsxSheetName, cellID, a)
		}
	} else {
		f := excelize.NewFile()
		if _, err := f.NewSheet(xlsxSheetName); err != nil {
			panic(err)
		}
		if err := f.SaveAs(xlsxPath); err != nil {
			panic(err)
		}
		closeXlsxFunc = func() {
			if err := f.SaveAs(xlsxPath); err != nil {
				panic(err)
			}
			f.Close()
		}
		newSheetHandler = func(a string) {
			if _, err := f.NewSheet(a); err != nil {
				panic(err)
			}
		}
		return func(line, col int, a string) error {
			cellID := fmt.Sprintf("%c%d", int8(col-1)+'A', line)
			return f.SetCellValue(xlsxSheetName, cellID, a)
		}
	}
}
