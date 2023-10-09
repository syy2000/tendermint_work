package randomsquare

import (
	"fmt"
	"testing"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
	"github.com/tendermint/tendermint/txgpartition/statustable"
	"github.com/xuri/excelize/v2"
)

var (
	GRAPHSIZE               = 60000
	_split_num              = 30
	_block_size             = 2000
	_tolerance_rate float64 = 0.1
	testTimes               = 10
)

var (
	writeXlsx       = true
	xlsxPath        = "./square_testout.xlsx"
	writeXlsxFunc   func(int, int, string) error
	closeXlsxFunc   func()
	newSheetHandler func(string)
	xlsxSheetName   = "Test"
)

func TestSquare(t *testing.T) {
	s := NewRandomSquare(GRAPHSIZE)
	s.RandomInit(0.5)
	fmt.Println("ok")
	start := time.Now()
	partitioning, cm, txMap := txgpartition.Init_Partitioning(s, _split_num, _tolerance_rate)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))
	fmt.Println("cut : ", txgpartition.CalculatePartitioningQualityByInnerPartitioningEdge(cm))

	// Simple Move
	start = time.Now()
	partitioning, cm, txMap = txgpartition.SimpleMove(s, _split_num, _tolerance_rate, partitioning, cm, txMap)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))

	// Advanced Move
	start = time.Now()
	_, cm, _ = txgpartition.AdvancedMove(s, _split_num, _tolerance_rate, partitioning, cm, txMap)
	fmt.Println("time used : ", time.Since(start))
	fmt.Println("numEdges", s.edgeNum, "numBLocks", s.blockNodeNum)
	fmt.Println("partition quality : ", txgpartition.CalculatePartitioningQualityByColorMap(cm))
}

func TestBatch(t *testing.T) {
	writeXlsxFunc = CreateOrOpen()
	defer closeXlsxFunc()
	for i := 40000; i <= 200000; i += 40000 {
		xlsxSheetName = fmt.Sprint(i)
		newSheetHandler(xlsxSheetName)
		writeXlsxFunc(2, 1, "time_used")
		writeXlsxFunc(3, 1, "cut")
		writeXlsxFunc(4, 1, "deps")
		writeXlsxFunc(5, 1, "edges")
		writeXlsxFunc(6, 1, "blockNums")
		writeXlsxFunc(1, 2, "init")
		writeXlsxFunc(1, 3, "sm")
		writeXlsxFunc(1, 4, "am")
		GRAPHSIZE, _split_num = i, i/_block_size
		var durI, durSM, durAM time.Duration
		var cut1, dep1 int = -1, -1
		var cut2, dep2 int = -1, -1
		var cut3, dep3 int = -1, -1
		g := NewRandomSquare(GRAPHSIZE)
		g.RandomInit(0.5)
		writeXlsxFunc(5, 2, fmt.Sprint(g.edgeNum))
		writeXlsxFunc(6, 2, fmt.Sprint(g.blockNodeNum))
		for ii := 0; ii < testTimes; ii++ {
			g.redo()
			ssstart := time.Now()
			fmt.Printf("Start test : GraphSize=%d, round=%d\n", GRAPHSIZE, ii)
			// init
			start := time.Now()
			partitioning, cm, txMap := txgpartition.Init_Partitioning(g, _split_num, _tolerance_rate)
			durI += time.Since(start)
			if out := txgpartition.CalculatePartitioningQualityByInnerPartitioningEdge(cm); cut1 >= 0 && cut1 != out {
				panic(fmt.Sprintf("non-deterministic cut: %d, %d", cut1, out))
			} else if cut1 == -1 {
				cut1 = out
			}
			if out := txgpartition.CalculatePartitioningQualityByColorMap(cm); dep1 >= 0 && dep1 != out {
				panic(fmt.Sprintf("non-deterministic dep: %d, %d", dep1, out))
			} else if dep1 == -1 {
				dep1 = out
			}
			// SM test
			start = time.Now()
			partitioning, cm, txMap = txgpartition.SimpleMove(g, _split_num, _tolerance_rate, partitioning, cm, txMap)
			durSM += time.Since(start)
			if out := txgpartition.CalculatePartitioningQualityByInnerPartitioningEdge(cm); cut2 >= 0 && cut2 != out {
				panic("non-deterministic cut")
			} else if cut2 == -1 {
				cut2 = out
			}
			if out := txgpartition.CalculatePartitioningQualityByColorMap(cm); dep2 >= 0 && dep2 != out {
				panic("non-deterministic cut")
			} else if dep2 == -1 {
				dep2 = out
			}
			// AM test
			start = time.Now()
			partitioning, cm, txMap = txgpartition.AdvancedMove(g, _split_num, _tolerance_rate, partitioning, cm, txMap)
			durAM += time.Since(start)
			if out := txgpartition.CalculatePartitioningQualityByInnerPartitioningEdge(cm); cut3 >= 0 && cut3 != out {
				panic("non-deterministic cut")
			} else if cut3 == -1 {
				cut3 = out
			}
			if out := txgpartition.CalculatePartitioningQualityByColorMap(cm); dep3 >= 0 && dep3 != out {
				panic("non-deterministic cut")
			} else if dep3 == -1 {
				dep3 = out
			}
			fmt.Printf("End  test : GraphSize=%d, round=%d in %s\n", GRAPHSIZE, ii, time.Since(ssstart))
		}
		durI /= time.Duration(testTimes)
		durSM /= time.Duration(testTimes)
		durAM /= time.Duration(testTimes)
		writeXlsxFunc(2, 2, fmt.Sprintf("%.2f", float64(durI)/float64(time.Millisecond)))
		writeXlsxFunc(3, 2, fmt.Sprint(cut1))
		writeXlsxFunc(4, 2, fmt.Sprint(dep1))
		writeXlsxFunc(2, 3, fmt.Sprintf("%.2f", float64(durSM)/float64(time.Millisecond)))
		writeXlsxFunc(3, 3, fmt.Sprint(cut2))
		writeXlsxFunc(4, 3, fmt.Sprint(dep2))
		writeXlsxFunc(2, 4, fmt.Sprintf("%.2f", float64(durAM)/float64(time.Millisecond)))
		writeXlsxFunc(3, 4, fmt.Sprint(cut3))
		writeXlsxFunc(4, 4, fmt.Sprint(dep3))
	}
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
