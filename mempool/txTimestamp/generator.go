package txTimestamp

import "github.com/tendermint/tendermint/types"

type Generator interface {
	SetSeed(seed *types.Seed) bool
	// 弃用
	// GenerateTimestamp(tx *types.Tx) types.TxTimestamp
	// 弃用
	SetOutputChan(out chan types.TxTimestamp)
	AddTx(tx types.TxWithTimestamp)
	GetTx() types.TxWithTimestamp

	// 状态调整查询
	GenStart() bool
	GenStop() bool
	GenQuery() bool
}
