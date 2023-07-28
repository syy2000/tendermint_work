package poH

import (
	"crypto/sha256"
	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
)

const (
	messageChanMaxNum = 10000
)

type PoHRound struct {
	Round   int64
	Message []byte
	PoH     []byte
}

type PoHGenerator struct {
	PoHRound     *PoHRound
	LastPoHRound *PoHRound

	service.BaseService

	LastTickRound int64
	IntervalRound int64

	mtx         tmsync.RWMutex
	MessageChan chan []byte

	OutChan chan types.TxTimestamp
}

func NewPoHGenerator(
	intervalRound int64,
	out chan types.TxTimestamp,
) *PoHGenerator {
	gen := new(PoHGenerator)
	gen.IntervalRound = intervalRound

	gen.MessageChan = make(chan []byte, messageChanMaxNum)
	gen.OutChan = out

	return gen
}

func (gen *PoHGenerator) SetOutputChan(out chan types.TxTimestamp) {
	gen.OutChan = out
}

// 生成下一个轮次
func (gen *PoHGenerator) generateNextRound(mes []byte) {
	h := sha256.New()
	h.Write(mes)
	h.Write(gen.PoHRound.PoH)

	nextRound := new(PoHRound)
	nextRound.Round = gen.PoHRound.Round + 1
	nextRound.PoH = h.Sum(nil)
	nextRound.Message = mes

	gen.LastPoHRound = gen.PoHRound
	gen.PoHRound = nextRound
}

// 返回当前PoH信息
func (gen *PoHGenerator) getPoHMessage() *types.PoHTimestamp {
	res := new(types.PoHTimestamp)
	res.Input = gen.LastPoHRound.PoH
	res.Out = gen.PoHRound.PoH
	res.Message = gen.PoHRound.Message
	res.Round = gen.PoHRound.Round

	return res
}

// Warning:该方法不是线程安全的
func (gen *PoHGenerator) GenerateTimestamp(tx *types.Tx) types.TxTimestamp {
	gen.generateNextRound(*tx)
	return gen.getPoHMessage()
}
