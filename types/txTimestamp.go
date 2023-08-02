package types

import (
	"github.com/tendermint/tendermint/crypto/merkle"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"strconv"
)

type TxTimestamp interface {
	// 返回int64时间戳，数字越小时间越早
	GetTimestamp() int64
	Hash() []byte
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

func (t *PoHTimestamp) Hash() []byte {
	leafs := make([][]byte, 4)
	leafs[0] = []byte(strconv.FormatInt(t.Round, 10))
	leafs[1] = t.Input
	leafs[2] = t.Message
	leafs[3] = t.Out
	return merkle.HashFromByteSlices(leafs)
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

type TimestampNormalError struct {
	err string
}

func (e *TimestampNormalError) Error() string {
	return e.err
}
