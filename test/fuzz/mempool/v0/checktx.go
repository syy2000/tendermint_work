package v0

import (
	"github.com/tendermint/tendermint/abci/example/kvstore"
	"github.com/tendermint/tendermint/config"
	mempl "github.com/tendermint/tendermint/mempool"
	mempoolv0 "github.com/tendermint/tendermint/mempool/v0"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
)

var mempool mempl.Mempool

func init() {
	app := kvstore.NewApplication()
	cc := proxy.NewLocalClientCreator(app)
	appConnMem, _ := cc.NewABCIClient()
	err := appConnMem.Start()
	if err != nil {
		panic(err)
	}

	cfg := config.DefaultMempoolConfig()
	cfg.Broadcast = false
	mempool = mempoolv0.NewCListMempool(cfg, appConnMem, 0)
}

func Fuzz(data types.Tx) int {
	//func Fuzz(data []byte) int {
	//Tx转MemTx
	err := mempool.CheckTx(data.OriginTx, nil, mempl.TxInfo{})
	if err != nil {
		return 0
	}

	return 1
}
