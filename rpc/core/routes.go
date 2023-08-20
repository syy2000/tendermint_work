package core

import (
	rpc "github.com/tendermint/tendermint/rpc/jsonrpc/server"
)

// TODO: better system than "unsafe" prefix

// Routes is a map of available routes.
var Routes = map[string]*rpc.RPCFunc{
	// subscribe/unsubscribe are reserved for websocket events.
	"subscribe":       rpc.NewWSRPCFunc(Subscribe, "query"),
	"unsubscribe":     rpc.NewWSRPCFunc(Unsubscribe, "query"),
	"unsubscribe_all": rpc.NewWSRPCFunc(UnsubscribeAll, ""),

	// info API
	"health":               rpc.NewRPCFunc(Health, ""),
	"status":               rpc.NewRPCFunc(Status, ""),
	"net_info":             rpc.NewRPCFunc(NetInfo, ""),
	"blockchain":           rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight", rpc.Cacheable()),
	"genesis":              rpc.NewRPCFunc(Genesis, "", rpc.Cacheable()),
	"genesis_chunked":      rpc.NewRPCFunc(GenesisChunked, "chunk", rpc.Cacheable()),
	"block":                rpc.NewRPCFunc(Block, "height", rpc.Cacheable("height")),
	"block_by_hash":        rpc.NewRPCFunc(BlockByHash, "hash", rpc.Cacheable()),
	"block_results":        rpc.NewRPCFunc(BlockResults, "height", rpc.Cacheable("height")),
	"commit":               rpc.NewRPCFunc(Commit, "height", rpc.Cacheable("height")),
	"check_tx":             rpc.NewRPCFunc(CheckTx, "tx"),
	"tx":                   rpc.NewRPCFunc(Tx, "hash,prove", rpc.Cacheable()),
	"tx_search":            rpc.NewRPCFunc(TxSearch, "query,prove,page,per_page,order_by"),
	"block_search":         rpc.NewRPCFunc(BlockSearch, "query,page,per_page,order_by"),
	"validators":           rpc.NewRPCFunc(Validators, "height,page,per_page", rpc.Cacheable("height")),
	"dump_consensus_state": rpc.NewRPCFunc(DumpConsensusState, ""),
	"consensus_state":      rpc.NewRPCFunc(ConsensusState, ""),
	"consensus_params":     rpc.NewRPCFunc(ConsensusParams, "height", rpc.Cacheable("height")),
	"unconfirmed_txs":      rpc.NewRPCFunc(UnconfirmedTxs, "limit"),
	"num_unconfirmed_txs":  rpc.NewRPCFunc(NumUnconfirmedTxs, ""),

	// tx broadcast API
	"broadcast_tx_commit": rpc.NewRPCFunc(BroadcastTxCommit, "tx"),
	"broadcast_tx_sync":   rpc.NewRPCFunc(BroadcastTxSync, "tx"),
	"broadcast_tx_async":  rpc.NewRPCFunc(BroadcastTxAsync, "tx"),

	// abci API
	"abci_query": rpc.NewRPCFunc(ABCIQuery, "path,data,height,prove"),
	"abci_info":  rpc.NewRPCFunc(ABCIInfo, "", rpc.Cacheable()),

	// evidence API
	"broadcast_evidence": rpc.NewRPCFunc(BroadcastEvidence, "evidence"),

	// 查询交易的上链情况
	"search_tx": rpc.NewRPCFunc(SearchTx, "tx,prove", rpc.Cacheable()),
	// 除了完成search_tx的工作之外，还查询了事务所在区块的提交时间。
	// 测试时，建议使用tx_time，因为它的功能覆盖面更广
	"tx_time": rpc.NewRPCFunc(TxTime, "tx,prove", rpc.Cacheable()),
	// 查询from到to之间区块的平均吞吐量
	"tps": rpc.NewRPCFunc(TpsCount, "from,to"),
	// 查询from到to之间区块的峰值吞吐量
	"tpsm": rpc.NewRPCFunc(TpsmCount, "from,to"),
}

// AddUnsafeRoutes adds unsafe routes.
func AddUnsafeRoutes() {
	// control API
	Routes["dial_seeds"] = rpc.NewRPCFunc(UnsafeDialSeeds, "seeds")
	Routes["dial_peers"] = rpc.NewRPCFunc(UnsafeDialPeers, "peers,persistent,unconditional,private")
	Routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(UnsafeFlushMempool, "")
}
