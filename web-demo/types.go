package main

import (
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"
)

type RPCResponse struct {
	JsonRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Result  ResultTime `json:"result"`
}

type ResultTime struct {
	Hash   bytes.HexBytes `json:"hash"`
	Height string         `json:"height"`
	Index  uint32         `json:"index"`
	Tx     types.Tx       `json:"tx"`
	Proof  types.TxProof  `json:"proof,omitempty"`
	Time   string         `json:"commit_block_time"`
	TPS    string         `json:"tps"`
}
