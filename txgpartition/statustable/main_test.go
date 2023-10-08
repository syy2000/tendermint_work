package statustable

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	testNameList = [][]string{
		{
			"donghao", "BIT", "SCZ", "url", "golang", "china", "waixingren", "taikongren", "woshishuaige", "shuaigeshiwo", "haloua", "ohayou",
		},
		{
			"8848", "2020", "13961164908", "15611396899", "15151214", "16161313150154", "1655555899", "151542125632", "11201961060", "3120230991", "1120191066", "1120192401", "1120192844",
		},
		{
			"d1h1d0h8", "plmi89875", "jkl236jj85", "8w5e8rr4y8u7u8i5", "y1h5j1k2f1s5s1", "dg858966545454", "58x9a22f2", "55f8c5a52sf5", "s8e4f", "lskdlgsj",
		},
		{
			"88ji69jk", "8985dghxd", "1x5d2s5e2c21g5d6", "2d5x2s41f54e52", "8d5c21s5g565", "5ca13fd546a3f1", "f5s4f65sd1", "c528se4f", "f418e4", "c5e984sf",
		},
		{
			"895dp65d2c", "5d21cd5g32", "djsncdgdlsndf", "sjdnvnsdjjsf45544141", "dsmmcndksk5541cd2s52", "54f5s241g2", "c5e29c5s", "f4fe84fs", "s544e98f4", "54cs8e4f,",
		},
	}
	maxTimes      = 100000 * 10
	testTimes     = 10
	testNumList   = make([]int64, maxTimes)
	testNumCtx    = false
	writeXlsx     = true
	xlsxPath      = "./testout.xlsx"
	writeXlsxFunc func(int, int, string) error
	closeXlsxFunc func()
	xlsxSheetName = "Test"
)

func TestMain(t *testing.T) {
	initNums()
	if writeXlsx {
		writeXlsxFunc = CreateOrOpen()
	}
	firstCols := []string{
		"mptHashSet", "mptHashGet", "mptHashChange", "mptHashHash",
		"mptSet", "mptGet", "mptChange", "mptHash",
		"simpleSet", "simpleGet", "simpleChange", "simpleHash",
		"orderSet", "orderGet", "orderChange", "orderHash",
	}
	for i, str := range firstCols {
		writeXlsxFunc(i+2, 1, str)
	}
	for i := 10; i <= 10; i++ {
		maxTimes = 100000 * i
		writeXlsxFunc(1, i+1, fmt.Sprint(maxTimes))
		iTestMPTHash(2, i+1)
		iTestMPT(6, i+1)
		iTestSimple(10, i+1)
		iTestOrder(14, i+1)
	}
	closeXlsxFunc()
}

func CreateOrOpen() func(int, int, string) error {
	if exists, err := PathExists(xlsxPath); err != nil {
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
		return func(line, col int, a string) error {
			cellID := fmt.Sprintf("%c%d", int8(col-1)+'A', line)
			return f.SetCellValue(xlsxSheetName, cellID, a)
		}
	}
}

func initNums() {
	if testNumCtx {
		return
	}
	defer func() { testNumCtx = true }()
	m := make(map[int64]bool)
	cnt := 0
	for cnt < maxTimes {
		n := rand.Int63()
		if !m[n] {
			m[n] = true
			testNumList[cnt] = n
			cnt++
		}
	}
}

func iTestKeys() ([]string, []int64) {
	keys, ids := []string{}, []int64{}
	for i := 0; i < len(testNameList); i++ {
		for j := 0; j < len(testNameList[i]); j++ {
			for u := 0; u < len(testNameList); u++ {
				for v := 0; v < len(testNameList[u]); v++ {
					key := testNameList[i][j] + testNameList[u][v]
					value := int64(i + j + u + v)
					keys = append(keys, key)
					ids = append(ids, value)
				}
			}
		}
	}
	return keys, ids
}

func iTestStore(m *BlockStatusMappingTable) (int, bool, time.Duration) {
	keys, ids := iTestKeys()
	cnt := len(keys)
	start := time.Now()
	for i := 0; i < cnt; i++ {
		if !m.Store(keys[i], ids[i]) {
			return 0, false, time.Duration(0)
		}
	}
	for i := 0; i < maxTimes-cnt; i++ {
		if !m.Store(fmt.Sprintf("%d", testNumList[i]), int64(i)) {
			return 0, false, time.Duration(0)
		}
	}
	return maxTimes, true, time.Since(start)
}
func iTestRead(m *BlockStatusMappingTable) (bool, time.Duration) {
	keys, ids := iTestKeys()
	cnt := len(keys)
	start := time.Now()
	for i := 0; i < cnt; i++ {
		value, ok := m.Load(keys[i])
		if !ok || value != ids[i] {
			return false, time.Duration(0)
		}
	}
	for i := 0; i < maxTimes-cnt; i++ {
		value, ok := m.Load(fmt.Sprintf("%d", testNumList[i]))
		if !ok || value != int64(i) {
			return false, time.Duration(0)
		}
	}
	return true, time.Since(start)
}

func iTestChange(m *BlockStatusMappingTable) (bool, time.Duration) {
	keys, ids := iTestKeys()
	cnt := len(keys)
	start := time.Now()
	for i := 0; i < cnt; i++ {
		if !m.Store(keys[i], ids[i]+1) {
			return false, time.Duration(0)
		}
	}
	for i := 0; i < maxTimes-cnt; i++ {
		if !m.Store(fmt.Sprintf("%d", testNumList[i]), int64(i+1)) {
			return false, time.Duration(0)
		}
	}
	return true, time.Since(start)
}

func iTestOnce(m *BlockStatusMappingTable, name string, line, col int) {
	fmt.Printf("============ %s test START ================ \n", name)

	initNums()
	var (
		cnt                                    int
		setTime, getTime, changeTime, hashTime time.Duration
		hash                                   []byte
	)
	for i := 0; i < testTimes; i++ {
		// set Test
		cntd, ok, setTimed := iTestStore(m)
		if !ok {
			panic("error: set failed")
		}
		cnt = cntd
		setTime += setTimed

		// get Test
		ok, getTimed := iTestRead(m)
		if !ok {
			panic("error: get failed")
		}
		getTime += getTimed

		// change Test
		ok, changeTimed := iTestChange(m)
		if !ok {
			panic("error: change failed")
		}
		changeTime += changeTimed

		// hash Test
		start := time.Now()
		hash = m.Hash()
		hashTime += time.Since(start)
		m.Clear()
	}
	setTime /= time.Duration(testTimes)
	getTime /= time.Duration(testTimes)
	changeTime /= time.Duration(testTimes)
	hashTime /= time.Duration(testTimes)
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println("set time : ", setTime)
	fmt.Println("get time : ", getTime)
	fmt.Println("change time : ", changeTime)
	fmt.Printf("hash : %x\n", hash)
	fmt.Println("hash time : ", hashTime)
	fmt.Printf("============ %s test END ================ \n\n", name)

	if writeXlsx {
		writeXlsxFunc(line, col, fmt.Sprintf("%.2f", float64(setTime.Microseconds())/float64(time.Millisecond.Microseconds())))
		writeXlsxFunc(line+1, col, fmt.Sprintf("%.2f", float64(getTime.Microseconds())/float64(time.Millisecond.Microseconds())))
		writeXlsxFunc(line+2, col, fmt.Sprintf("%.2f", float64(changeTime.Microseconds())/float64(time.Millisecond.Microseconds())))
		writeXlsxFunc(line+3, col, fmt.Sprintf("%.2f", float64(hashTime.Microseconds())/float64(time.Millisecond.Microseconds())))
	}
}

func iTestMPTHash(line, col int) {
	m := NewBlockStatusMappingTable(UseMPTree, OnlyUseHashOptions)
	iTestOnce(m, "MPTHash", line, col)
}

func iTestMPT(line, col int) {
	m := NewBlockStatusMappingTable(UseMPTree, nil)
	iTestOnce(m, "MPT", line, col)
}

func iTestOrder(line, col int) {
	m := NewBlockStatusMappingTable(UseOrderedMap, nil)
	iTestOnce(m, "OrderedMap", line, col)
}

func iTestSimple(line, col int) {
	m := NewBlockStatusMappingTable(UseSimpleMap, nil)
	iTestOnce(m, "SimpleMap", line, col)
}

func iTestSafeSimple(line, col int) {
	m := NewBlockStatusMappingTable(UseSafeSimpleMap, nil)
	iTestOnce(m, "SafeSimpleMap", line, col)

}
