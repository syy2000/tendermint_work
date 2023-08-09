package txTimestamp

import "github.com/tendermint/tendermint/types"

type Generator interface {
	SetSeed(seed *types.Seed) bool

	AddTx(tx types.TxWithTimestamp)
	GetTx(id int64) types.TxWithTimestamp

	// 状态调整查询
	GenStart() bool
	GenStop() bool
	GenQuery() bool
}
