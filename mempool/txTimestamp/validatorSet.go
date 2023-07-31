package txTimestamp

import "github.com/tendermint/tendermint/types"

type ValidatorSet interface {
	AddValidator(validator Validator) bool
	RemoveValidator(validator Validator) bool
	GetTxs() []types.TxWithTimestamp
}
