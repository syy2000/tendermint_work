package poH

import (
	"crypto/sha256"
	"sync"

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

	mtx tmsync.RWMutex
	// tx输入
	MessageChan chan types.TxWithTimestamp

	// 打好时间戳的tx输出
	TxOutChan chan types.TxWithTimestamp

	mempool *PoHMempool

	flag bool
	quit chan struct{}

	Logger log.Logger

	txWithTimestampMap sync.Map
	// map[int64]txxypes.TxWithTimestamp
	mtx2 tmsync.Mutex
}

func NewPoHGenerator(
	intervalRound int64,
	log log.Logger,
	mempool *PoHMempool,
) *PoHGenerator {
	gen := new(PoHGenerator)
	gen.IntervalRound = intervalRound

	gen.MessageChan = make(chan types.TxWithTimestamp, MessageChanMaxNum)
	gen.TxOutChan = make(chan types.TxWithTimestamp, MessageChanMaxNum)
	gen.mempool = mempool

	gen.PoHRound = nil
	gen.flag = false
	gen.quit = make(chan struct{}, 5)
	gen.Logger = log
	gen.txWithTimestampMap = sync.Map{}
	return gen
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
	gen.generateNextRound(tx.OriginTx)
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
		gen.LastTickRound = gen.PoHRound.Round

		res := gen.getPoHMessage()
		//gen.OutChan <- res
		gen.mempool.AddTimestamp(res)
		if txOutFlag {
			tx.SetTimestamp(res)
			gen.TxOutChan <- tx
			// gen.txWithTimestampMap[tx.GetId()] = tx
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
func (gen *PoHGenerator) GetTx(id int64) types.TxWithTimestamp {
	tx, ok := gen.txWithTimestampMap.Load(id)
	if !ok {
		gen.mtx2.Lock()
		defer gen.mtx2.Unlock()
		tx, ok := gen.txWithTimestampMap.Load(id)
		if ok {
			gen.txWithTimestampMap.Delete(id)
			return tx.(types.TxWithTimestamp)
		}
		for {
			select {
			case tx = <-gen.TxOutChan:
				temp := tx.(types.TxWithTimestamp)
				if temp.GetId() == id {
					return temp
				}
				gen.txWithTimestampMap.Store(temp.GetId(), temp)
			}
		}
	}
	gen.txWithTimestampMap.Delete(id)
	return tx.(types.TxWithTimestamp)
}

func (gen *PoHGenerator) GetTxChan() chan types.TxWithTimestamp {
	return gen.TxOutChan
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
