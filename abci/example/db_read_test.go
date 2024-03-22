package example

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	//"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	charSet2 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func generateUniqueStrings2(keyNum int) ([][]byte, [][]byte) {
	visited := make(map[string]bool)
	results := make([][]byte, 0, keyNum)
	for len(results) < keyNum {
		b := [4]byte{} // 只生成4个字节的key
		for i := range b {
			b[i] = charSet2[rand.Intn(len(charSet2))]
		}
		str := string(b[:])
		if !visited[str] {
			visited[str] = true
			results = append(results, []byte(str))
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return bytes.Compare(results[i], results[j]) < 0
	})
	return results, results
}

func TestLevelDBRead(t *testing.T) {
	db, err := leveldb.OpenFile("testdb_read", &opt.Options{})
	if err != nil {
		if errors.IsCorrupted(err) {
			db, err = leveldb.RecoverFile("testdb", nil)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	defer db.Close()
	keyNum := 1000000
	//write_ops := 1000
	key := make([][]byte, keyNum)
	value := make([][]byte, keyNum)
	newValue := []byte("new_value")
	key, value = generateUniqueStrings2(keyNum)
	for i := 0; i < keyNum; i++ {
		err := db.Put(key[i], value[i], nil)
		if err != nil {
			panic("写入数据时出错：" + err.Error())
		}
		//fmt.Printf("%q %q\n", key[i], value[i])
	}

	//********************读入开始********************************
	//随机写
	// for epoch := 0; epoch < 9; epoch++ {
	// 	write_ops := [9]int{100, 200, 300, 400, 500, 600, 700, 800, 900}
	// 	write_op := write_ops[epoch]
	// 	key, value = generateUniqueStrings2(keyNum)
	// 	//times := 10
	// 	start := time.Now()
	// 	for i := 0; i < write_op; i++ {
	// 		err := db.Put(key[rand.Intn(keyNum)], newValue, nil)
	// 		if err != nil {
	// 			panic("读数据出错：" + err.Error())
	// 		}
	// 	}
	// 	elapsed := time.Since(start)
	// 	//fmt.Printf("read ops = %d  read throughput: %.2f ops/sec\n", read_ops, float64(read_ops)/(elapsed.Seconds()))
	// 	fmt.Printf("%.2f\n", float64(write_op)/(elapsed.Seconds()))
	// }
	//顺序写
	// for epoch := 0; epoch < 9; epoch++ {
	// 	write_ops := [9]int{100, 200, 300, 400, 500, 600, 700, 800, 900}
	// 	write_op := write_ops[epoch]
	// 	//times := 10
	// 	key, value = generateUniqueStrings2(keyNum)
	// 	start := time.Now()
	// 	for i := 0; i < write_op; i++ {
	// 		err := db.Put(key[i%keyNum], newValue, nil)
	// 		if err != nil {
	// 			panic("读数据出错：" + err.Error())
	// 		}
	// 	}
	// 	elapsed := time.Since(start)
	// 	//fmt.Printf("read ops = %d  read throughput: %.2f ops/sec\n", read_ops, float64(read_ops)/(elapsed.Seconds()))
	// 	fmt.Printf("%.2f\n", float64(write_op)/(elapsed.Seconds()))
	// }
	//重复写
	for epoch := 0; epoch < 9; epoch++ {
		write_ops := [9]int{100, 200, 300, 400, 500, 600, 700, 800, 900}
		write_op := write_ops[epoch]
		key, value = generateUniqueStrings2(keyNum)

		//times := 10
		tmp := rand.Intn(keyNum)
		start := time.Now()
		for i := 0; i < write_op; i++ {
			err := db.Put(key[tmp], newValue, nil)
			if err != nil {
				panic("读数据出错：" + err.Error())
			}
		}
		elapsed := time.Since(start)
		//fmt.Printf("read ops = %d  read throughput: %.2f ops/sec\n", read_ops, float64(read_ops)/(elapsed.Seconds()))
		fmt.Printf("%.2f\n", float64(write_op)/(elapsed.Seconds()))
	}
}
