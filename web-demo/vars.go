package main

import (
	"bufio"
	"os"
	"strconv"
)

const (
	latencyTest_txs_per_second int    = 15
	latencyTest_total_time     int    = 10
	latencyTest_resp_or_commit string = "commit"
	latencyTest_routines       int    = 10
	latencyTest_pay_num        int    = 1
	latencyTest_get_num        int    = 1
	latencyTest_file_path      string = "latencytestConfig.txt"
)

const (
	workloadTest_pay_num     = 2
	workLoadTest_get_num     = 2
	workloadTest_hot_rate    = 0.1
	workloadTest_money_range = 10000
	loadtestPrefix           = "workload:"
	getTPSPrefix             = "tps:"
)

var (
	latencyTest_pay_money_min int = 5000
	latencyTest_pay_money_max int = 6000
)

func InitLatency() {
	if _, err := os.Stat(latencyTest_file_path); err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(latencyTest_file_path)
			if err != nil {
				panic(err)
			}
			_, err = f.WriteString(strconv.Itoa(workloadTest_money_range))
			if err != nil {
				f.Close()
				panic(err)
			}
			f.Close()
		} else {
			panic(err)
		}
	}
	f, err := os.OpenFile(latencyTest_file_path, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ff := bufio.NewReader(f)
	s, _, _ := ff.ReadLine()
	u, err := strconv.Atoi(string(s))
	if err != nil {
		panic(err)
	}
	latencyTest_pay_money_min = u
	latencyTest_pay_money_max = u + latencyTest_txs_per_second*latencyTest_total_time*5

	f2, err := os.OpenFile(latencyTest_file_path, os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	_, err = f2.WriteString(strconv.Itoa(latencyTest_pay_money_max))
	if err != nil {
		panic(err)
	}
}

func CalculateRespTime() bool {
	return latencyTest_resp_or_commit == "resp"
}
func CalculateCommitTime() bool {
	return latencyTest_resp_or_commit == "commit"
}
