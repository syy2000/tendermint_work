package txTimestamp

import (
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)

type TxState interface {
	// 需要更准确的peer标识
	AddValidator(validator *types.Validator) bool
	RemoveValidator(peerID p2p.ID) bool
	GetTxChan() chan types.TxWithTimestamp
	AddMessage(*types.TxMessage)

	SetSeed(seed *types.Seed) bool

	Start() error

	GetNowTimestamp() int64
}
