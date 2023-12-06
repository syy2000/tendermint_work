package scoin

import (
	"fmt"
	"testing"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
)

func TestCreateCoin(t *testing.T) {
	txInsert := CreateInsertTx("donghao", 15)
	txBuy := CreateTransferTx([]string{"donghao", "qiuxiaopeng"}, []string{"lyz"}, 2)
	txi := types.Tx{OriginTx: txInsert}
	txb := types.Tx{OriginTx: txBuy}

	app := NewApplication("")
	reqInsert := abcitypes.RequestCheckTx{Tx: txi.ToProto()}
	reqBuy := abcitypes.RequestCheckTx{Tx: txb.ToProto()}
	reqInsertDeliver := abcitypes.RequestDeliverTx{Tx: txi.ToProto()}
	reqBuyDeliver := abcitypes.RequestDeliverTx{Tx: txb.ToProto()}
	//fmt.Println(app.GetDB())
	fmt.Println(RWAnalyse(txi))
	fmt.Println(RWAnalyse(txb))
	fmt.Println(app.CheckTx(reqInsert))
	fmt.Println(app.CheckTx(reqBuy))
	fmt.Println(app.DeliverTx(reqInsertDeliver))
	fmt.Println(app.DeliverTx(reqBuyDeliver))
	fmt.Println(app)
}
