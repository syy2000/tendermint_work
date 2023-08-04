package poH

import (
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
	"sync/atomic"
)

const (
	TxNumOneBlock      = 5000
	TxTimestampChanMax = 100000
	CreateBlockChanMax = 100
)

type PoHMempool struct {
	// 输入
	TxTimestampChan chan types.TxTimestamp

	// 可以打包的信号
	CreateBlockChan chan struct{}
	txNum           int64
	Height          int64
	Round           int64
	Logger          log.Logger
	mtx             sync.Mutex
}

func NewPoHMempool(
	logger log.Logger) *PoHMempool {
	m := new(PoHMempool)
	m.TxTimestampChan = make(chan types.TxTimestamp, TxTimestampChanMax)
	m.CreateBlockChan = make(chan struct{}, CreateBlockChanMax)
	m.txNum = 0
	m.Height = 0
	m.Round = 0
	m.Logger = logger
	return m
}

func (m *PoHMempool) GetTimestampChan() chan types.TxTimestamp {
	return m.TxTimestampChan
}

func (m *PoHMempool) GetCreateBlockChan() chan struct{} {
	return m.CreateBlockChan
}

func (m *PoHMempool) SetSeed(seed *types.Seed) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.Height = seed.Height
	m.Round = seed.Round
}

// TODO 判断是否创建时是按照固定大小来的，等待细化
func (m *PoHMempool) AddTimestamp(t types.TxTimestamp) {
	m.TxTimestampChan <- t
	num := atomic.AddInt64(&m.txNum, 1)
	if num%TxNumOneBlock == 0 {
		// 正好到整倍数
		m.CreateBlockChan <- struct{}{}
	}
}

// TODO 目前获取是固定大小的
func (m *PoHMempool) GetTimestamps() []*types.PoHTimestamp {
	l := TxNumOneBlock
	res := make([]*types.PoHTimestamp, l)

	for i := 0; i < l; i++ {
		t := <-m.TxTimestampChan
		temp := t.(*types.PoHTimestamp)
		res[i] = temp
	}
	atomic.AddInt64(&m.txNum, -int64(l))
	atomic.AddInt64(&m.Height, 1)
	return res
}
