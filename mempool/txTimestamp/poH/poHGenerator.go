package poH

import (
	"crypto/sha256"

	"github.com/tendermint/tendermint/libs/log"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
)

const (
	MessageChanMaxNum = 10000
	roundMax          = 100000000000000000
	roundMin          = 0
	roundInterval     = 1000000
)

type PoHRound struct {
	Round   int64
	Message []byte
	PoH     []byte
}

type PoHGenerator struct {
	PoHRound     *PoHRound
	LastPoHRound *PoHRound

	//service.BaseService

	LastTickRound int64
	IntervalRound int64

	mtx         tmsync.RWMutex
	MessageChan chan types.TxWithTimestamp
	OutChan     chan types.TxTimestamp
	TxOutChan   chan types.TxWithTimestamp

	flag bool
	quit chan struct{}

	Logger log.Logger
}

func NewPoHGenerator(
	intervalRound int64,
	out chan types.TxTimestamp,
	log log.Logger,
) *PoHGenerator {
	gen := new(PoHGenerator)
	gen.IntervalRound = intervalRound

	gen.MessageChan = make(chan types.TxWithTimestamp, MessageChanMaxNum)
	gen.TxOutChan = make(chan types.TxWithTimestamp, MessageChanMaxNum)
	gen.OutChan = out

	gen.PoHRound = nil
	gen.flag = false
	gen.quit = make(chan struct{}, 5)
	gen.Logger = log
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
// 弃用
func (gen *PoHGenerator) GenerateTimestamp(tx *types.Tx) types.TxTimestamp {
	gen.generateNextRound(*tx)
	return gen.getPoHMessage()
}

// 这里目前是使用异步减少阻塞
func (gen *PoHGenerator) AddTx(tx types.TxWithTimestamp) {
	go func() {
		gen.MessageChan <- tx
	}()
}

func (gen *PoHGenerator) generateNextRoundAndOutput() {
	mes := make([]byte, 0)
	var tx types.TxWithTimestamp

	tickFlag := false
	txOutFlag := false
	select {
	case tx = <-gen.MessageChan:
		mes = tx.GetTx()
		txOutFlag = true
	default:
	}
	if gen.PoHRound.Round-gen.LastTickRound >= gen.IntervalRound {
		tickFlag = true
	}
	gen.generateNextRound(mes)
	if txOutFlag || tickFlag {
		res := gen.getPoHMessage()
		gen.OutChan <- res
		if txOutFlag {
			tx.SetTimestamp(res)
			gen.TxOutChan <- tx
		}
	}
}

func (gen *PoHGenerator) generate() {
	gen.Logger.Debug("生成启动")
	for {
		select {
		case <-gen.quit:
			gen.mtx.Lock()
			gen.flag = false
			gen.mtx.Unlock()
			return
		default:
			gen.generateNextRoundAndOutput()
		}
	}
}

func (gen *PoHGenerator) SetSeed(seed *types.Seed) bool {
	gen.mtx.RLock()
	defer gen.mtx.RUnlock()
	if gen.flag {
		return false
	}
	poHRound := new(PoHRound)

	SeedCopy := make([]byte, len(seed.Seed))
	copy(SeedCopy, seed.Seed)
	poHRound.PoH = SeedCopy
	poHRound.Message = make([]byte, 0)
	poHRound.Round = seed.Round

	gen.PoHRound = poHRound
	return true
}

// 同步函数
func (gen *PoHGenerator) GetTx() types.TxWithTimestamp {
	return <-gen.TxOutChan
}

func (gen *PoHGenerator) GenStart() bool {
	gen.mtx.Lock()
	defer gen.mtx.Unlock()
	if gen.flag {
		return true
	}
	gen.flag = true
	go gen.generate()

	return true
}

func (gen *PoHGenerator) GenStop() bool {
	gen.mtx.RLock()
	defer gen.mtx.RUnlock()
	if gen.flag {
		gen.quit <- struct{}{}
		return true
	}
	return false
}

func (gen *PoHGenerator) GenQuery() bool {
	gen.mtx.RLock()
	defer gen.mtx.RUnlock()
	return gen.flag
}
