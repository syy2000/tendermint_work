package txTimestamp

import "github.com/tendermint/tendermint/types"

type Validator interface {
	SetSeed(seed *types.Seed)

	GetNowTimestamp() types.TxTimestamp

	GetNextBlockHeight() int64

	Validate(block types.TxBlock) bool
}
