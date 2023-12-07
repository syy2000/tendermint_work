package v0

import (
	"fmt"
	"math"
	"math/rand"
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
	step = 40000
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
	for i:=0; i<accountNum; i++ {
		accountMap.Store(accountNum, 10000)
	}
	for i := 1; i <= 1  ; i++ {
		total := step * i
		fmt.Printf("========= Node Num : %d ==========\n", total)
		writeXlsxFunc(1, i+1, fmt.Sprint(total))
		var (
			time_used                                 time.Duration
			edges, zero_outdegree, max_deps, mid_deps, count, totalWeight int
			Sequential_total_time, Concurrent_total_time float64
			//Concurrent_total_time float64
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
			time_used += time.Since(start)
			edges += mem.edgeNum()
			zero_outdegree = len(mem.FindZeroOutdegree()) - preBlockNum
			max_deps = mem.maxDep()
			mid_deps = mem.midDep() 
			count = mem.countComponent()
			totalWeight += int(mem.countWeight())
			Sequential_total_time += mem.ExecuteSequentially(accountMap)
			mem.ExecuteConcurrently(accountMap)
			Concurrent_total_time += mem.ExecuteConcurrently(accountMap)
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
	}
}

func GenReadWrites(n int) []*mempoolTx {
	out := make([]*mempoolTx, n)
	for i := 0; i < n; i++ {
		var weight int64
		memTx := types.MemTx{
			TxId:        int64(i),
			TxOp:        []string{},
			TxObAndAttr: []string{},
		}
		keys := GenRandomKey()
		for j := 0; j < len(keys); j++ {
			if j < len(keys)/2 {
				memTx.TxOp = append(memTx.TxOp, "read")
				weight += 1
			} else {
				memTx.TxOp = append(memTx.TxOp, "write")
				weight += 2
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
func TestRandomKey(t *testing.T){
	keys := GenRandomKey()
	for i:=0; i<len(keys); i++ {
		fmt.Printf("%s", keys[i])
	}
}
func GenRandomKey() []string {
	outInt := []int{}
	for len(outInt) < keyNumPerTx {
		x := rand.Intn(accountNum)
		flg := false
		for _, y := range outInt {
			if x == y {
				flg = true
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
		out[i] = fmt.Sprint(x)
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

