package txTimestamp

import "github.com/tendermint/tendermint/types"

type Generator interface {
	SetSeed(seed *types.Seed)
	GenerateTimestamp(tx *types.Tx) types.TxTimestamp
	SetOutputChan(out chan types.TxTimestamp)

	// 状态调整查询
	GenStart() bool
	GenStop() bool
	GenQuery() bool
}
