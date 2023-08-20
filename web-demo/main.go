package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	coin "github.com/tendermint/tendermint/abci/example/scoin"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var keys []ed25519.PrivKey
var pubkeys []string

var IP = []string{
	"127.0.0.1:26657",
	"127.0.0.1:36657",
	"127.0.0.1:46657",
	"127.0.0.1:56657",
}

func main() {

	num := len(os.Args)
	if num == 1 {
		panic("need a path! insert-all  latency  or demo")
	}
	if num == 2 {
		panic("we need at least 1 IP!")
	}
	IP = os.Args[2:]
	path := os.Args[1]

	fmt.Println("Init......")
	get_keys()
	fmt.Println("Finish!")

	if path == "insert-all" {
		insertAllLocal(10000000)
		return
	} else if path == "latency" {
		latencyTest()
		return
	} else if path == "demo" {
		fmt.Println("IP list : ")
		for _, ip := range IP {
			fmt.Println(ip)
		}
	} else if strings.HasPrefix(path, loadtestPrefix) {
		workLoadTest(path[len(loadtestPrefix):])
		return
	} else if strings.HasPrefix(path, getTPSPrefix) {
		TpsCount(path[len(getTPSPrefix):])
		return
	} else {
		panic("need a path : insert-all  latency  or demo")
	}

	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/tx", txHandler)
	http.HandleFunc("/insertall", insertAll)
	http.HandleFunc("/query_tx", queryTxHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	var thekey = r.FormValue("key")
	var themoney = r.FormValue("money")
	if len(thekey) == 0 {
		fmt.Fprintf(w, "No Content!\n")
		return
	}
	keyid, err := strconv.Atoi(thekey)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	money, err := strconv.Atoi(themoney)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	if keyid >= len(keys) || keyid < 0 {
		fmt.Fprintln(w, "Want an 0 <= id < 1000")
		return
	}

	tx := MarshalInsert(pubkeys[keyid], money)

	url := fmt.Sprintf("http://%s/broadcast_tx_commit?tx=\"%s\"", IP[rand.Intn(len(IP))], tx)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	} else {
		var out []byte
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		fmt.Fprintln(w, string(out))
		resp.Body.Close()
		return
	}

}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	var thekey = r.FormValue("key")
	if len(thekey) == 0 {
		fmt.Fprintf(w, "No Content!\n")
		return
	}
	keyid, err := strconv.Atoi(thekey)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	tx := fmt.Sprintf("http://%s/abci_query?data=\"%s\"", IP[rand.Intn(len(IP))], pubkeys[keyid])
	resp, err := http.Get(tx)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	} else {
		var out []byte
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		fmt.Fprintln(w, string(out))
		resp.Body.Close()
		return
	}
}

func get_keys() {
	f, err := os.Open("generate_keys/keys.txt")
	if err != nil {
		fmt.Println(err)
		return
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
}

func txHandler(w http.ResponseWriter, r *http.Request) {
	sz1 := r.FormValue("from")
	sz2 := r.FormValue("to")
	themoney := r.FormValue("money")

	fromstrs := strings.Split(sz1, "!")
	tostrs := strings.Split(sz2, "!")

	money, err := strconv.Atoi(themoney)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	var froms, tos []string

	for _, s := range fromstrs {
		id, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		froms = append(froms, pubkeys[id])
	}

	for _, s := range tostrs {
		id, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		tos = append(tos, pubkeys[id])
	}

	tx := coin.CreateTransferTx(froms, tos, int32(money))

	fmt.Println(string(tx))
	url := fmt.Sprintf("http://%s/broadcast_tx_commit?tx=\"%s\"", IP[rand.Intn(len(IP))], tx)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	} else {
		var out []byte
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		fmt.Fprintln(w, string(out))
		resp.Body.Close()
		return
	}

}

func queryTxHandler(w http.ResponseWriter, r *http.Request) {
	sz1 := r.FormValue("from")
	sz2 := r.FormValue("to")
	themoney := r.FormValue("money")

	fromstrs := strings.Split(sz1, "!")
	tostrs := strings.Split(sz2, "!")

	money, err := strconv.Atoi(themoney)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	var from, to = make([]int, len(fromstrs)), make([]int, len(tostrs))
	for i, s := range fromstrs {
		tmp, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		from[i] = tmp
	}
	for i, s := range tostrs {
		tmp, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		to[i] = tmp
	}

	tx := generateTxPay(from, to, money)
	url := fmt.Sprintf("http://%s/tx_time?tx=\"%s\"&prove=true", IP[rand.Intn(len(IP))], string(tx))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	} else {
		var out []byte
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		fmt.Fprintln(w, string(out))
		resp.Body.Close()
		var rsp RPCResponse
		if err := json.Unmarshal(out, &rsp); err != nil {
			tstr := rsp.Result.Time
			tunix, _ := strconv.ParseInt(tstr, 10, 64)
			t := time.UnixMicro(tunix)
			fmt.Println(t)
		} else {
			fmt.Println(err)
		}
		return
	}

}
func MarshalInsert(pubkey string, money int) []byte {
	return coin.CreateInsertTx(pubkey, int32(money))
}

func insertAll(w http.ResponseWriter, r *http.Request) {
	moneystr := r.FormValue("money")
	money, err := strconv.Atoi(moneystr)
	if err != nil {
		fmt.Fprintln(w, err)
	} else {
		mod := 85
		taskCh := make(chan int, mod*len(IP))
		backCh := make(chan bool, mod*len(IP))
		for i := 0; i < mod*len(IP); i++ {
			go func(k int) {
				for id := range taskCh {
					tx := coin.CreateInsertTx(pubkeys[id], int32(money))
					var theip string = IP[k/mod]
					url := fmt.Sprintf("http://%s/broadcast_tx_async?tx=\"%s\"", theip, tx)
					rsp, err := http.Get(url)
					if err != nil {
						fmt.Println(err)
					}
					rsp.Body.Close()
					backCh <- err != nil
				}
			}(i)
		}
		go func() {
			for i := 0; i < len(pubkeys); i++ {
				taskCh <- i
			}
		}()
		for i := 0; i < len(pubkeys); i++ {
			<-backCh
			if (i+1)%100 == 0 {
				fmt.Println((i+1)/10, "%")
			}
		}
		close(taskCh)
		close(backCh)
		fmt.Fprintln(w, "ok")
	}
}

func insertAllLocal(money int) {
	var err error = nil
	if err != nil {
		return
	} else {
		mod := 85
		fmt.Println(mod * len(IP))
		taskCh := make(chan int, mod*len(IP))
		backCh := make(chan bool, mod*len(IP))
		for i := 0; i < mod*len(IP); i++ {
			go func(k int) {
				for id := range taskCh {
					tx := coin.CreateInsertTx(pubkeys[id], int32(money))
					var theip string = IP[k/mod]
					url := fmt.Sprintf("http://%s/broadcast_tx_async?tx=\"%s\"", theip, tx)
					rsp, err := http.Get(url)
					if err != nil {
						fmt.Println(err)
					}
					rsp.Body.Close()
					backCh <- err != nil
				}
			}(i)
		}
		go func() {
			for i := 0; i < len(pubkeys); i++ {
				taskCh <- i
			}
		}()
		for i := 0; i < len(pubkeys); i++ {
			ot := <-backCh
			if (i+1)%100 == 0 {
				fmt.Println(ot, (i+1)/10, "%")
			}
		}
		close(taskCh)
		close(backCh)
	}
}

func latencyTest() {
	InitLatency()

	total := latencyTest_txs_per_second * latencyTest_total_time
	clientTime := make([]time.Time, total)
	txs := make([][]byte, total)
	respTime := make([]time.Duration, total)
	commitTime := make([]time.Duration, total)
	// 产生交易
	generateTxRoutine := func(start int) {
		for i := start; i < total; i += latencyTest_routines {
			from := rand.Intn(len(keys))
			to := rand.Intn(len(keys))
			for to == from {
				to = rand.Intn(len(keys))
			}
			txs[i] = generateTxPay([]int{from}, []int{to}, rand.Intn(latencyTest_pay_money_max-latencyTest_pay_money_min)+latencyTest_pay_money_min)
		}
	}
	for i := 0; i < latencyTest_routines; i++ {
		go generateTxRoutine(i)
	}

	// 发送交易；commit方法
	wg := sync.WaitGroup{}
	wg.Add(total)
	tick := time.NewTicker(1 * time.Second / time.Duration(latencyTest_txs_per_second))
	commitTxRoutine := func(id int) {
		defer wg.Done()
		var theip string = IP[rand.Intn(len(IP))]
		clientTime[id] = time.Now()
		url := fmt.Sprintf("http://%s/broadcast_tx_commit?tx=\"%s\"", theip, string(txs[id]))
		http.Get(url)
		respTime[id] = time.Since(clientTime[id])
	}
	for i := 0; i < total; i++ {
		go commitTxRoutine(i)
		<-tick.C
	}
	wg.Wait()

	if CalculateRespTime() {
		var avatime time.Duration = 0
		for _, resp_time := range respTime {
			avatime += resp_time
		}
		fmt.Printf("latency : %f ms\n", float64(avatime)/float64(total)/float64(time.Millisecond))
	} else if CalculateCommitTime() {

		var avatime time.Duration = 0
		var cnt int
		wg.Add(total)

		getTxCommitTime := func(id int) {
			defer wg.Done()
			var theip string = IP[rand.Intn(len(IP))]
			url := fmt.Sprintf("http://%s/tx_time?tx=\"%s\"", theip, string(txs[id]))
			resp, err := http.Get(url)
			if err != nil {
				panic(err)
			}
			out, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			var rsp RPCResponse
			json.Unmarshal(out, &rsp)
			if timeUnix, _ := strconv.ParseInt(rsp.Result.Time, 10, 64); timeUnix > 0 {
				commitTime[id] = time.UnixMicro(timeUnix).Sub(clientTime[id])
			}
		}

		for i := 0; i < total; i++ {
			go getTxCommitTime(i)
		}
		wg.Wait()
		for _, cmtTime := range commitTime {
			if cmtTime > 0 {
				avatime += cmtTime
				cnt++
			}
		}
		fmt.Printf("latency : %f ms\n", float64(avatime)/float64(cnt)/float64(time.Millisecond))
		return
	} else {
		panic("vars.go : latencyTest_resp_or_commit must be 'resp' or 'commit'")
	}

}

func generateTxPay(pays, gets []int, money int) []byte {
	var from, to []string
	for _, id := range pays {
		from = append(from, pubkeys[id])
	}
	for _, id := range gets {
		to = append(to, pubkeys[id])
	}
	return coin.CreateTransferTx(from, to, int32(money))
}

func workLoadTest(args string) {
	input_args := strings.Split(args, "+")
	m := map[string]int{}
	for _, arg := range input_args {
		k := strings.Split(arg, "=")
		t, err := strconv.Atoi(k[1])
		if err != nil {
			panic(err)
		}
		m[k[0]] = t
	}
	load, ok1 := m["load"]
	totalTime, ok2 := m["time"]
	num_conn, ok3 := m["conn"]
	ok := ok1 && ok2 && ok3
	if !ok {
		panic(fmt.Sprintf("need argument like %sload=200+time=60+conn=5\n", loadtestPrefix))
	}

	f, err := os.Open(fmt.Sprintf("generate_keys/txs-%d-%d.txt", workLoadTest_get_num, workloadTest_pay_num))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ch := make(chan *string, num_conn*100)
	defer close(ch)

	mtx := sync.Mutex{}
	var success int
	var wg = sync.WaitGroup{}
	wg.Add(num_conn)
	deliverRoutine := func(id int) {
		var (
			innerSuccess, roundSuccess int
			round                      int
		)
		start := time.Now()
		for tx := range ch {
			url := fmt.Sprintf("http://%s/broadcast_tx_async?tx=\"%s\"", randIP(), *tx)
			_, err := http.Get(url)
			if err == nil {
				roundSuccess++
			}
			if time.Since(start) > time.Second {
				start = time.Now()
				innerSuccess += roundSuccess
				fmt.Printf("thread %d round %d : timeout with %d txs delivered to RPC\n", id, round, roundSuccess)
				roundSuccess = 0
				round++
			} else if roundSuccess == load {
				innerSuccess += roundSuccess
				fmt.Printf("thread %d round %d : success with %d txs\n", id, round, roundSuccess)
				roundSuccess = 0
				round++
				time.Sleep(time.Second - time.Since(start))
				start = time.Now()
			}
			if round == totalTime {
				break
			}
		}
		mtx.Lock()
		success += innerSuccess
		mtx.Unlock()
		wg.Done()
	}
	for i := 0; i < num_conn; i++ {
		go deliverRoutine(i)
	}

	h := bufio.NewReader(f)
	for i := 0; i < totalTime*num_conn*load; i++ {
		out, _, _ := h.ReadLine()
		s := string(out)
		ch <- &s
	}
	wg.Wait()
	fmt.Println("total : ", success)

}

func randIP() string {
	return IP[rand.Intn(len(IP))]
}

/*
func modIP(id int) string {
	return IP[id%len(IP)]
}
*/

func TpsCount(msg string) {
	us := strings.Split(msg, "-")
	if len(us) != 2 {
		panic("error TpsCount! for example:" + getTPSPrefix + "5-25")
	}
	from, to := us[0], us[1]
	var theip string = randIP()
	url := fmt.Sprintf("http://%s/tps?from=\"%s\"&to=\"%s\"", theip, from, to)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	var rsp RPCResponse
	json.Unmarshal(out, &rsp)
	tm, _ := strconv.ParseFloat(rsp.Result.TPS, 64)
	fmt.Printf("height [%s,%s] : %f tps\n", from, to, tm)
}
