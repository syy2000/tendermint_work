package types

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type TxTimestamp interface {
	// 返回int64时间戳，数字越小时间越早
	GetTimestamp() int64
}

type PoHTimestamp struct {
	Round   int64
	Input   []byte
	Message []byte
	Out     []byte
}

func (t *PoHTimestamp) GetTimestamp() int64 {
	return t.Round
}

func (t *PoHTimestamp) ToProto() *tmproto.PoHTimestamp {
	if t == nil {
		return nil
	}
	return &tmproto.PoHTimestamp{
		Round:   t.Round,
		Input:   t.Input,
		Message: t.Message,
		Out:     t.Out,
	}
}

func NewPoHTimestampFromProto(pt *tmproto.PoHTimestamp) *PoHTimestamp {
	if pt == nil {
		return nil
	}
	return &PoHTimestamp{
		Round:   pt.Round,
		Input:   pt.Input,
		Message: pt.Message,
		Out:     pt.Out,
	}
}

type Seed struct {
	Seed   []byte
	Height int64
	Round  int64
}

type TxWithTimestamp interface {
	GetTx() []byte
	SetTimestamp(t TxTimestamp)
	GetTimestamp() TxTimestamp
}
