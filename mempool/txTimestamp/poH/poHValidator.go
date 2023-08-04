package poH

import (
	"bytes"
	"crypto/sha256"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

type PoHValidator struct {
	Height        int64
	Round         int64
	lastPoH       []byte
	lastTimestamp *types.PoHTimestamp

	Logger log.Logger
}

func NewPoHValidator(logger log.Logger) *PoHValidator {
	res := new(PoHValidator)
	res.Height = 0
	res.Round = 0
	res.lastPoH = make([]byte, 0)
	res.lastTimestamp = nil
	res.Logger = logger
	return res
}

func (v *PoHValidator) SetSeed(seed *types.Seed) {
	v.Logger.Debug("设置种子", "height", v.Height)
	v.lastPoH = make([]byte, len(seed.Seed))
	copy(v.lastPoH, seed.Seed)

	v.lastTimestamp = new(types.PoHTimestamp)
	v.lastTimestamp.Input = make([]byte, 0)
	v.lastTimestamp.Out = seed.Seed
	v.lastTimestamp.Round = seed.Round

	v.Height = seed.Height
	v.Round = seed.Round
}

func (v *PoHValidator) GetNowTimestamp() types.TxTimestamp {
	return v.lastTimestamp
}

func (v *PoHValidator) GetNextBlockHeight() int64 {
	return v.Height
}

func (v *PoHValidator) Validate(block types.TxBlock) bool {
	b, err := block.(*types.PoHBlock)
	if !err {
		return false
	}
	if !b.BaseValidate() {
		return false
	}
	if b.Height != v.Height {
		return false
	}
	f, _ := v.ValidateBlock(b)
	if !f {
		return false
	}
	v.Height++
	return true
}

func (v *PoHValidator) ValidateMes(m *types.PoHTimestamp) bool {
	h := sha256.New()
	h.Write(m.Message)
	h.Write(m.Input)
	res := h.Sum(nil)
	return bytes.Compare(res, m.Out) == 0
}

func (v *PoHValidator) ValidateTick(poHStart []byte, startRound int64, poHEnd []byte, endRound int64) bool {
	now := poHStart
	for i := startRound; i < endRound; i++ {
		h := sha256.New()
		h.Write(now)
		now = h.Sum(nil)
	}
	return bytes.Compare(now, poHEnd) == 0
}

func (v *PoHValidator) ValidateBlock(b *types.PoHBlock) (bool, int) {
	messageNum := 0
	num := len(b.PoHTimestamps) * 2
	res := make(chan bool, num)
	go func() {
		res <- v.ValidateTick(v.lastPoH, v.Round, b.PoHTimestamps[0].Input, b.PoHTimestamps[0].Round-1)
	}()
	for i := range b.PoHTimestamps {
		go func(j int) {
			res <- v.ValidateMes(b.PoHTimestamps[j])
		}(i)
		if i != 0 {
			go func(j int) {
				res <- v.ValidateTick(b.PoHTimestamps[j-1].Out, b.PoHTimestamps[j-1].Round, b.PoHTimestamps[j].Input, b.PoHTimestamps[j].Round-1)
			}(i)
		}
	}
	for i := 0; i < num; i++ {
		flag := <-res
		if !flag {
			return false, 0
		}
	}
	return true, messageNum
}
