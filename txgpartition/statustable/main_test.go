package statustable

import (
	"fmt"
	"testing"
	"time"
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
	maxTimes  = 1000000
	readTimes = 5
)

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
		if !m.Store(fmt.Sprintf("%d", i), int64(i)) {
			return 0, false, time.Duration(0)
		}
	}
	return maxTimes, true, time.Since(start)
}
func iTestRead(m *BlockStatusMappingTable) (bool, time.Duration) {
	keys, ids := iTestKeys()
	cnt := len(keys)
	start := time.Now()
	for ii := 0; ii < readTimes; ii++ {
		for i := 0; i < cnt; i++ {
			value, ok := m.Load(keys[i])
			if !ok || value != ids[i] {
				return false, time.Duration(0)
			}
		}
		for i := 0; i < maxTimes-cnt; i++ {
			value, ok := m.Load(fmt.Sprintf("%d", i))
			if !ok || value != int64(i) {
				return false, time.Duration(0)
			}
		}
	}
	return true, time.Since(start)
}

func iTestOnce(m *BlockStatusMappingTable, name string) {
	fmt.Printf("============ %s test START ================ \n", name)

	// set Test
	cnt, ok, setTime := iTestStore(m)
	if !ok {
		panic("error: get failed")
	}
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println("set time : ", setTime)

	// get Test
	ok, getTime := iTestRead(m)
	if !ok {
		panic("error: get failed")
	}
	fmt.Println("get time : ", getTime)

	// hash Test
	start := time.Now()
	fmt.Printf("hash : %x\n", m.Hash())
	fmt.Println("hash time : ", time.Since(start))
	fmt.Printf("============ %s test END ================ \n\n", name)
}

func TestMPT(t *testing.T) {
	m := NewBlockStatusMappingTable(UseMPTree, OnlyUseHashOptions)
	iTestOnce(m, "MPT")
}

func TestOrder(t *testing.T) {
	m := NewBlockStatusMappingTable(UseOrderedMap, nil)
	iTestOnce(m, "OrderedMap")
}

func TestSimple(t *testing.T) {
	m := NewBlockStatusMappingTable(UseSimpleMap, nil)
	iTestOnce(m, "SimpleMap")
}

func TestMain(t *testing.T) {
	m := NewBlockStatusMappingTable(UseSafeSimpleMap, nil)
	iTestOnce(m, "SafeSimpleMap")

}
