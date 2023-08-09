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
)

func TestMPT(t *testing.T) {
	fmt.Println("MPT test : ")
	mpt := NewMPT()
	mpt.useHash = true
	cnt := 0
	start := time.Now()
	for i := 0; i < len(testNameList); i++ {
		for j := 0; j < len(testNameList[i]); j++ {
			for u := 0; u < len(testNameList); u++ {
				for v := 0; v < len(testNameList[u]); v++ {
					key := testNameList[i][j] + testNameList[u][v]
					value := int64(i + j + u + v)
					mpt.set(key, blockHeight(value))
					cnt++
				}
			}
		}
	}
	for i := 0; i < 1000000; i++ {
		mpt.set(fmt.Sprintf("%d", i), blockHeight(i))
		cnt++
	}
	fmt.Println("total time : ", time.Since(start))
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println(mpt.get("donghaodonghao"))
	start = time.Now()
	fmt.Printf("hash : %x\n", mpt.hash())
	fmt.Println("hash time : ", time.Since(start))
	fmt.Println("=====================================")
}

func TestOrder(t *testing.T) {
	fmt.Println("OrderedMap test : ")
	mpt := NewOrderedMap()
	cnt := 0
	start := time.Now()
	for i := 0; i < len(testNameList); i++ {
		for j := 0; j < len(testNameList[i]); j++ {
			for u := 0; u < len(testNameList); u++ {
				for v := 0; v < len(testNameList[u]); v++ {
					key := testNameList[i][j] + testNameList[u][v]
					value := int64(i + j + u + v)
					mpt.set(key, blockHeight(value))
					cnt++
				}
			}
		}
	}
	for i := 0; i < 1000000; i++ {
		mpt.set(fmt.Sprintf("%d", i), blockHeight(i))
		cnt++
	}
	fmt.Println("total time : ", time.Since(start))
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println(mpt.get("donghaodonghao"))
	start = time.Now()
	fmt.Printf("hash : %x\n", mpt.hash())
	fmt.Println("hash time : ", time.Since(start))
	fmt.Println("=====================================")
}

func TestSimple(t *testing.T) {
	fmt.Println("simple test : ")
	mpt := NewSimpleMap()
	cnt := 0
	start := time.Now()
	for i := 0; i < len(testNameList); i++ {
		for j := 0; j < len(testNameList[i]); j++ {
			for u := 0; u < len(testNameList); u++ {
				for v := 0; v < len(testNameList[u]); v++ {
					key := testNameList[i][j] + testNameList[u][v]
					value := int64(i + j + u + v)
					mpt.set(key, blockHeight(value))
					cnt++
				}
			}
		}
	}
	for i := 0; i < 1000000; i++ {
		mpt.set(fmt.Sprintf("%d", i), blockHeight(i))
		cnt++
	}
	fmt.Println("total time : ", time.Since(start))
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println(mpt.get("donghaodonghao"))
	start = time.Now()
	fmt.Printf("hash : %x\n", mpt.hash())
	fmt.Println("hash time : ", time.Since(start))
	fmt.Println("=====================================")
}

func TestMain(t *testing.T) {
	fmt.Println("test main use safe simple map: ")
	mpt := NewBlockStatusMappingTable(UseSafeSimpleMap, nil)
	cnt := 0
	start := time.Now()
	for i := 0; i < len(testNameList); i++ {
		for j := 0; j < len(testNameList[i]); j++ {
			for u := 0; u < len(testNameList); u++ {
				for v := 0; v < len(testNameList[u]); v++ {
					key := testNameList[i][j] + testNameList[u][v]
					value := int64(i + j + u + v)
					mpt.Store(key, int64(value))
					cnt++
				}
			}
		}
	}
	for i := 0; i < 1000000; i++ {
		mpt.Store(fmt.Sprintf("%d", i), int64(i))
		cnt++
	}
	fmt.Println("total time : ", time.Since(start))
	fmt.Printf("calcutate %d times\n", cnt)
	fmt.Println(mpt.Load("donghaodonghao"))
	start = time.Now()
	fmt.Printf("hash : %x\n", mpt.Hash())
	fmt.Println("hash time : ", time.Since(start))
	fmt.Println("=====================================")
}
