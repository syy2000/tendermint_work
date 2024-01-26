package kvstore

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	//"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
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
	keyLength := 16
	keyNum := 1000000
	//readTimes := 10000
	writeTimes := 1000
	key, value := genRandomKeys(keyNum, keyLength)
	for i := 0; i < keyNum; i++ {
		err := db.Put(key[i], value[i], nil)
		if err != nil {
			panic("写入数据时出错：" + err.Error())
		}
		//fmt.Printf("%q %q\n", key[i], value[i])
	}
	//fmt.Println("写入成功")
	fmt.Println("Way of key generation: random")
	fmt.Printf("keyNum: %d\n", keyNum)
	fmt.Printf("keyLength: %d\n", keyLength)
	//fmt.Printf("readTimes: %d\n", readTimes)
	fmt.Printf("writeTimes: %d\n", writeTimes)
	// // 读入测试
	// //多次读取随机生成的index,key[index]
	// start := time.Now()
	// for time := 0; time < 10; time++ {
	// 	for i := 0; i < readTimes; i++ {
	// 		_, err := db.Get(key[rand.Intn(keyNum)], nil)
	// 		if err != nil {
	// 			panic("读数据出错：" + err.Error())
	// 		}
	// 	}
	// }
	// elapsed := time.Since(start)
	// fmt.Printf("Read randomly throughput(average): %.2f ops/sec\n", float64(readTimes*10)/(elapsed.Seconds()))
	// //顺序读取
	// start = time.Now()
	// for time := 0; time < 10; time++ {
	// 	for i := 0; i < readTimes; i++ {
	// 		_, err := db.Get(key[i], nil)
	// 		if err != nil {
	// 			panic("读数据出错：" + err.Error())
	// 		}
	// 	}
	// }
	// elapsed = time.Since(start)
	// fmt.Printf("Read sequently throughput(average): %.2f ops/sec\n", float64(readTimes*10)/elapsed.Seconds())
	// // 重复读取
	// tmp := rand.Intn(keyNum)
	// start = time.Now()
	// for time := 0; time < 10; time++ {
	// 	for i := 0; i < readTimes; i++ {
	// 		_, err := db.Get(key[tmp], nil)
	// 		if err != nil {
	// 			panic("读数据出错：" + err.Error())
	// 		}
	// 	}
	// }
	// elapsed = time.Since(start)
	// fmt.Printf("Read repeatly throughput(average): %.2f ops/sec\n", float64(readTimes*10)/elapsed.Seconds())

	// 写入测试
	//多次写随机生成的index,key[index]
	start := time.Now()
	newValue := []byte("new_value")
	for time := 0; time < 10; time++ {
		for i := 0; i < writeTimes; i++ {
			err := db.Put(key[rand.Intn(keyNum)], newValue, nil)
			if err != nil {
				panic("写数据出错：" + err.Error())
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Write randomly throughput(average): %.2f ops/sec\n", float64(writeTimes*10)/(elapsed.Seconds()))
	//顺序写入
	start = time.Now()
	for time := 0; time < 10; time++ {
		for i := 0; i < writeTimes; i++ {
			err := db.Put(key[i], newValue, nil)
			if err != nil {
				panic("写数据出错：" + err.Error())
			}
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("Write sequently throughput(average): %.2f ops/sec\n", float64(writeTimes*10)/elapsed.Seconds())
	// 重复写入
	tmp := rand.Intn(keyNum)
	start = time.Now()
	for time := 0; time < 10; time++ {
		for i := 0; i < writeTimes; i++ {
			err := db.Put(key[tmp], newValue, nil)
			if err != nil {
				panic("写数据出错：" + err.Error())
			}
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("Write repeatly throughput(average): %.2f ops/sec\n", float64(writeTimes*10)/elapsed.Seconds())
}
