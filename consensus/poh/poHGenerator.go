package poh

import (
	"crypto/sha256"

	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
)

const (
	MESSAGECHANMAX    = 100000
	POHMESSAGECHANMAX = 100000
)

// 第Round轮次，得到了消息Message，结果为PoH
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
	mtx           tmsync.RWMutex
	MessageChan   chan []byte

	PoHMesOutChan   chan *types.PoHMessage
	LeaderStartFlag chan struct{}
	leaderFlag      bool
}

func NewPoHGenerator(
	intervalRound int64,
) *PoHGenerator {
	gen := new(PoHGenerator)
	gen.IntervalRound = intervalRound
	gen.MessageChan = make(chan []byte, MESSAGECHANMAX)
	gen.PoHMesOutChan = make(chan *types.PoHMessage, POHMESSAGECHANMAX)

	gen.LeaderStartFlag = make(chan struct{}, 5)
	gen.leaderFlag = false

	gen.BaseService = *service.NewBaseService(nil, "PoHGenerator", gen)
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
func (gen *PoHGenerator) GetPoHMessage() *types.PoHMessage {
	res := new(types.PoHMessage)
	res.Input = gen.LastPoHRound.PoH
	res.Out = gen.PoHRound.PoH
	res.Message = gen.PoHRound.Message
	res.Round = gen.PoHRound.Round

	return res
}

func (gen *PoHGenerator) generate() {
	for {
		select {
		case <-gen.Quit():
			return
		case <-gen.LeaderStartFlag:
			for gen.IsLeader() {
				select {
				case <-gen.Quit():
					return
				default:
					gen.generateNextRoundAndOutput()
				}
			}
		}
	}
}

func (gen *PoHGenerator) IsLeader() bool {
	gen.mtx.RLock()
	defer gen.mtx.RUnlock()
	return gen.leaderFlag
}

func (gen *PoHGenerator) SetLeader(flag bool) {
	gen.mtx.Lock()
	defer gen.mtx.Unlock()
	gen.leaderFlag = flag
}

func (gen *PoHGenerator) generateNextRoundAndOutput() {
	mes := make([]byte, 0)
	flag := false
	select {
	case mes = <-gen.MessageChan:
		flag = true
	default:
	}
	if gen.PoHRound.Round-gen.LastTickRound >= gen.IntervalRound {
		flag = true
	}
	gen.generateNextRound(mes)
	if flag {
		//TODO: 这里是输出位置
		res := gen.GetPoHMessage()
		gen.PoHMesOutChan <- res
		gen.LastTickRound = gen.LastPoHRound.Round
	}
}

func (gen *PoHGenerator) OnStart() error {
	gen.Logger.Debug("pohGen 启动中")
	go func() {
		gen.generate()
	}()
	return nil
}

func (gen *PoHGenerator) SetRound(round *PoHRound) {
	gen.PoHRound = round

	gen.LastTickRound = round.Round
}
