package poH

import (
	"bytes"
	"fmt"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
	"testing"
	"time"
)

type TestTx struct {
	Tx []byte
	t  types.TxTimestamp
}

func (tx *TestTx) GetTx() []byte {
	return tx.Tx
}

func (tx *TestTx) SetTimestamp(t types.TxTimestamp) {
	tx.t = t
}

func (tx *TestTx) GetTimestamp() types.TxTimestamp {
	return tx.t
}

func pohTimestampDeepEqual(t1 *types.PoHTimestamp, t2 *types.PoHTimestamp) bool {
	return bytes.Equal(t1.Input, t2.Input) && bytes.Equal(t1.Message, t2.Message) && bytes.Equal(t1.Out, t2.Out) && t1.Round == t2.Round
}

func TestPoHGen(t *testing.T) {

	logger := log.NewNopLogger()
	mempool := NewPoHMempool(logger)
	gen := NewPoHGenerator(10000, logger, mempool)

	seed := &types.Seed{
		Seed:   []byte("hello world"),
		Height: 10,
		Round:  1000,
	}
	// val.SetSeed(seed)
	//privKey := ed25519.GenPrivKey()
	//pubKey := privKey.PubKey()
	//address := pubKey.Address()
	//state := NewPoHTxState(mempool, gen, privKey, pubKey, address)
	//
	//state.Start()
	//
	//val := &types.Validator{
	//	Address:          address,
	//	PubKey:           pubKey,
	//	VotingPower:      0,
	//	ProposerPriority: 0,
	//}
	//state.AddValidator(val)
	//
	//state.SetSeed(seed)
	gen.SetSeed(seed)
	txs := []*TestTx{
		&TestTx{Tx: []byte("hello"), t: nil},
		&TestTx{Tx: []byte("world"), t: nil},
		&TestTx{Tx: []byte("!"), t: nil},
	}

	gen.GenStart()
	for _, tx := range txs {
		gen.AddTx(tx)
	}
	outTxs := make([]*TestTx, 0)
	outTxs = append(outTxs, gen.GetTx().(*TestTx))
	outTxs = append(outTxs, gen.GetTx().(*TestTx))
	outTxs = append(outTxs, gen.GetTx().(*TestTx))

	gen.GenStop()

	outTxsByOrder := make([]*TestTx, 3)
	for _, tx := range outTxs {
		if bytes.Equal(tx.Tx, []byte("hello")) {
			outTxsByOrder[0] = tx
		}
		if bytes.Equal(tx.Tx, []byte("world")) {
			outTxsByOrder[1] = tx
		}
		if bytes.Equal(tx.Tx, []byte("!")) {
			outTxsByOrder[2] = tx
		}
	}

	for i := range txs {
		if !bytes.Equal(txs[i].Tx, outTxsByOrder[i].Tx) {
			t.Error("输出的tx不等")
		}
		time := outTxsByOrder[i].t.(*types.PoHTimestamp)
		if !bytes.Equal(txs[i].Tx, time.Message) {
			t.Error("时间戳内含的tx出错啦")
		}
	}

	mempoolTxs := make([]*types.PoHTimestamp, 0)
	num := mempool.txNum
	fmt.Printf("mempool的num = %v\n", num)
	for i := int64(0); i < num; i++ {
		temp := (<-mempool.TxTimestampChan).(*types.PoHTimestamp)
		if len(temp.Message) != 0 {
			mempoolTxs = append(mempoolTxs, temp)
		}
	}

	if len(mempoolTxs) != len(outTxs) {
		t.Error("输出到mempool的tx缺失")
	}
	for i := 0; i < len(outTxs); i++ {
		t1 := (outTxs[i].t).(*types.PoHTimestamp)
		t2 := mempoolTxs[i]
		if !pohTimestampDeepEqual(t1, t2) {
			t.Error("输出到mempool的tx与返回的tx时间戳不同！")
		}
	}

}

func TestPoHState(t *testing.T) {
	logger := log.NewNopLogger()
	mempool := NewPoHMempool(logger)
	gen := NewPoHGenerator(3000, logger, mempool)
	seed := &types.Seed{
		Seed:   []byte("hello world"),
		Height: 10,
		Round:  1000,
	}
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	address := pubKey.Address()
	state := NewPoHTxState(mempool, gen, privKey, pubKey, address)

	state.Start()

	val := &types.Validator{
		Address:          address,
		PubKey:           pubKey,
		VotingPower:      0,
		ProposerPriority: 0,
	}
	state.AddValidator(val)

	state.SetSeed(seed)

	txs := []*TestTx{
		&TestTx{Tx: []byte("hello"), t: nil},
		&TestTx{Tx: []byte("world"), t: nil},
		&TestTx{Tx: []byte("!"), t: nil},
	}

	gen.GenStart()
	for _, tx := range txs {
		gen.AddTx(tx)
	}

	select {
	case <-time.After(3 * time.Second):
	}
	gen.GenStop()
	for flag := true; flag; {
		select {
		case ps := <-state.OutPoHBlockPartSetChan:
			for i := uint32(0); i < ps.Total(); i++ {
				state.MessageChan <- &types.TxMessage{
					Src:  p2p.ID(address),
					Data: ps.GetPart(int(i)),
				}
			}
		default:
			flag = false
			break
		}
	}
	//select {
	//case <-time.After(1 * time.Second):
	//}
	//bTxs := make([]*types.PoHTimestamp, 3)
	//for flag := true; flag; {
	//	select {
	//	case tx := <-state.testPoHTimestamp:
	//		if len(tx.Message) != 0 {
	//			if bytes.Equal(tx.Message, []byte("hello")) {
	//				bTxs[0] = tx
	//			}
	//			if bytes.Equal(tx.Message, []byte("world")) {
	//				bTxs[1] = tx
	//			}
	//			if bytes.Equal(tx.Message, []byte("!")) {
	//				bTxs[2] = tx
	//			}
	//		}
	//	default:
	//		flag = false
	//		break
	//	}
	//}
	//
	//for i := range txs {
	//	if !bytes.Equal(txs[i].Tx, bTxs[i].Message) {
	//		t.Error("输出的tx不等")
	//	}
	//}
}
