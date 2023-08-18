package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	num_to_generate         = 10000         // 生成的账户数量
	payNum                  = 2             // 测试集中每笔交易支付账户的个数
	totalNum                = 4             // 测试集中涉及的总个数
	generate_keys           = true          // 是否重新生成账户
	generate_txs            = false         // 是否重新生成交易
	keys_path               = "keys.txt"    // 账户数据保存地址，不需要修改
	txs_path                = "txs-2-2.txt" // 交易数据保存地址，不需要修改
	txs_to_generate         = 300000        // 生成交易数量
	hot_rate        float64 = 0.05          // 热门账户占比
	hot_num                 = int(math.Floor(hot_rate * float64(num_to_generate)))
)

func TestGenerateKeys(t *testing.T) {
	if !generate_keys {
		return
	}

	f, err := os.OpenFile(keys_path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < num_to_generate; i++ {
		privKey := createKey()
		_, err = f.WriteString(hex.EncodeToString(privKey.Bytes()) + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func TestGenerateTxs(t *testing.T) {
	start := time.Now()
	if !generate_txs {
		return
	}
	keys, pubkeys := get_keys()

	f, err := os.OpenFile(txs_path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	clock := 10
	epoch := txs_to_generate / 100
	for i := 0; i < txs_to_generate; i++ {
		tx, _ := GenerateTx(keys, pubkeys)
		_, err = f.WriteString(string(tx) + "\n")
		if err != nil {
			panic(err)
		}
		if i == clock*epoch {
			fmt.Printf("generate tx : %d %% ;\n", clock)
			clock += 10
		}
	}

	dur := float64(time.Since(start)) / float64(time.Second)
	fmt.Printf("successfully generate %d txs in %f s, %f tps\n", txs_to_generate, dur, float64(txs_to_generate)/dur)

}

//===========================================================================================================

func createKey() ed25519.PrivKey {
	return ed25519.GenPrivKey()
}

func getPubKey(privkey ed25519.PrivKey) ed25519.PubKey {
	return privkey.PubKey().Bytes()
}

type Insert struct {
	Flag   int8
	Pubkey []byte
	Money  int32
}

type Buy struct {
	Flag  int8
	From  [][]byte
	To    [][]byte
	Sigs  [][]byte
	Money int32
}

func generateRandomNumber(start int, end int, hot int, rate float64, count int) []int {
	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		var num int
		if r.Float64() > rate {
			num = r.Intn((hot - start)) + start
		} else {
			num = r.Intn((end - hot)) + hot
		}

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

func get_keys() (keys []ed25519.PrivKey, pubkeys []string) {
	f, err := os.Open(keys_path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	h := bufio.NewScanner(f)
	for h.Scan() {
		s, err := hex.DecodeString(h.Text())
		if err != nil {
			panic(err)
		}
		privkey := ed25519.PrivKey(s)
		keys = append(keys, privkey)
		pubkeys = append(pubkeys, hex.EncodeToString(privkey.PubKey().Bytes()))
	}
	return
}

func GenerateTx(privKeys []ed25519.PrivKey, pubkeys []string) ([]byte, error) {

	var accounts []int
	num := len(pubkeys)
	msg := []byte("ok")
	var buy = Buy{
		From:  make([][]byte, payNum),
		To:    make([][]byte, totalNum-payNum),
		Flag:  2,
		Sigs:  make([][]byte, payNum),
		Money: rand.Int31n(10000) + 1,
	}

	accounts = generateRandomNumber(0, num, hot_num, hot_rate, totalNum)
	rand.Shuffle(totalNum, func(i, j int) {
		accounts[i], accounts[j] = accounts[j], accounts[i]
	})

	for i := 0; i < payNum; i++ {
		id := accounts[i]
		buy.From[i] = []byte(pubkeys[id])
		s, err := privKeys[id].Sign(msg)
		if err != nil {
			return nil, err
		} else {
			buy.Sigs[i] = []byte(hex.EncodeToString(s))
		}
	}
	for i := 0; i < totalNum-payNum; i++ {
		id := accounts[i+payNum]
		buy.To[i] = []byte(pubkeys[id])
	}

	o, err := json.Marshal(buy)
	if err != nil {
		return nil, err
	}
	for i, bt := range o {
		if bt == '"' {
			o[i] = '$'
		}
	}

	return o, nil
}
