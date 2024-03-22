package kvstore

import (
	"bytes"
	"fmt"
	"math"
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
	charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func genRandomKeys(keyNum int, keyLength int) ([][]byte, [][]byte) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	src := []byte(letterBytes)
	key := make([][]byte, keyNum)
	value := make([][]byte, keyNum)
	for i := 0; i < keyNum; i++ {
		key[i] = make([]byte, keyLength)
		value[i] = make([]byte, keyLength)
		for j := 0; j < keyLength; j++ {
			key[i][j] = src[rand.Intn(len(src))]
			value[i][j] = src[rand.Intn(len(src))]
		}
	}
	return key, value
}
func generateUniqueStrings(keyNum int) ([][]byte, [][]byte) {
	visited := make(map[string]bool)
	results := make([][]byte, 0, keyNum)
	for len(results) < keyNum {
		b := [4]byte{} // 只生成4个字节的key
		for i := range b {
			b[i] = charSet[rand.Intn(len(charSet))]
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
func TestLevelDB(t *testing.T) {

	db, err := leveldb.OpenFile("testdb", &opt.Options{})
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

	//start := time.Now()
	//keyLength := 16
	keyNum := 10000
	op := 10000
	read_rate := 0.1
	repeat_rate := 0.1
	// repeatTimes := int(math.Ceil(float64(op) * repeat_rate))
	tmp := rand.Intn(keyNum)
	newValue := []byte("new_value")
	//key, value := genRandomKeys(keyNum, keyLength)
	key := make([][]byte, keyNum)
	value := make([][]byte, keyNum)
	key, value = generateUniqueStrings(keyNum)
	for i := 0; i < keyNum; i++ {
		err := db.Put(key[i], value[i], nil)
		if err != nil {
			panic("写入数据时出错：" + err.Error())
		}
		//fmt.Printf("%q %q\n", key[i], value[i])
	}
	fmt.Println("写入成功")
	fmt.Printf("keyNum: %d\n", keyNum)

	// 读入测试
	// 多次读随机生成的键
	for epoch := 1; epoch <= 10; epoch++ {
		read_rate = float64(epoch) * 0.1
		readTimes := int(math.Ceil(float64(op) * read_rate))
		writeTimes := op - readTimes
		for k := 1; k <= 10; k++ {
			repeat_rate = float64(k) * 0.1
			//重复读
			readRepeat := int(math.Ceil(float64(readTimes) * repeat_rate))
			writeRepeat := int(math.Ceil(float64(writeTimes) * repeat_rate))
			start := time.Now()
			for i := 0; i < readRepeat; i++ {
				_, err := db.Get(key[tmp], nil)
				if err != nil {
					panic("读数据出错：" + err.Error())
				}
			}
			//剩下的读
			for i := readRepeat; i < readTimes; i++ {
				_, err := db.Get(key[rand.Intn(keyNum)], nil)
				if err != nil {
					panic("读数据出错：" + err.Error())

				}
			}
			//重复写
			for i := 0; i < writeRepeat; i++ {
				err := db.Put(key[tmp], newValue, nil)
				if err != nil {
					panic("写数据出错：" + err.Error())
				}
			}
			// 剩下的写
			for i := writeRepeat; i < writeTimes; i++ {
				err := db.Put(key[rand.Intn(keyNum)], newValue, nil)
				if err != nil {
					panic("写数据出错：" + err.Error())
				}
			}
			elapsed := time.Since(start)
			fmt.Printf("%d\n", elapsed.Nanoseconds())
		}
		//fmt.Printf("****************************************************************\n")
		//fmt.Printf("%.2f\n", float64(writeTimes)/(elapsed.Seconds()))
	}

}
