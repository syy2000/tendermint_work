package core

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/tendermint/tendermint/state/txindex/null"
	"github.com/tendermint/tendermint/types"
)

// Tx allows you to query the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place.
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/tx
func Tx(ctx *rpctypes.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, fmt.Errorf("transaction indexing is disabled")
	}

	r, err := env.TxIndexer.Get(hash)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	height := r.Height
	index := r.Index

	var proof types.TxProof
	if prove {
		block := env.BlockStore.LoadBlock(height)
		proof = block.Data.Txs.Proof(int(index)) // XXX: overflow on 32-bit machines
	}

	return &ctypes.ResultTx{
		Hash:     hash,
		Height:   height,
		Index:    index,
		TxResult: r.Result,
		Tx:       *types.NewTxFromProto(r.Tx),
		Proof:    proof,
	}, nil
}

// TxSearch allows you to query for multiple transactions results. It returns a
// list of transactions (maximum ?per_page entries) and the total count.
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/tx_search
func TxSearch(
	ctx *rpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*ctypes.ResultTxSearch, error) {

	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, errors.New("transaction indexing is disabled")
	} else if len(query) > maxQueryLength {
		return nil, errors.New("maximum query length exceeded")
	}

	q, err := tmquery.New(query)
	if err != nil {
		return nil, err
	}

	results, err := env.TxIndexer.Search(ctx.Context(), q)
	if err != nil {
		return nil, err
	}

	// sort results (must be done before pagination)
	switch orderBy {
	case "desc":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index > results[j].Index
			}
			return results[i].Height > results[j].Height
		})
	case "asc", "":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index < results[j].Index
			}
			return results[i].Height < results[j].Height
		})
	default:
		return nil, errors.New("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := validatePerPage(perPagePtr)

	page, err := validatePage(pagePtr, perPage, totalCount)
	if err != nil {
		return nil, err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*ctypes.ResultTx, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		r := results[i]

		var proof types.TxProof
		if prove {
			block := env.BlockStore.LoadBlock(r.Height)
			proof = block.Data.Txs.Proof(int(r.Index)) // XXX: overflow on 32-bit machines
		}

		apiResults = append(apiResults, &ctypes.ResultTx{
			Hash:     types.Tx(*types.NewTxFromProto(r.Tx)).Hash(),
			Height:   r.Height,
			Index:    r.Index,
			TxResult: r.Result,
			Tx:       *types.NewTxFromProto(r.Tx),
			Proof:    proof,
		})
	}

	return &ctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount}, nil
}

// 查询交易的上链情况
func SearchTx(ctx *rpctypes.Context, tx []byte, prove bool) (*ctypes.ResultTx, error) {
	utx := types.Tx{
		OriginTx: tx,
	}
	hash := utx.Hash()
	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, fmt.Errorf("transaction indexing is disabled")
	}
	// 先通过index服务查询交易所在区块、交易执行结果等信息
	r, err := env.TxIndexer.Get(hash)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	var proof types.TxProof
	if prove {
		block := env.BlockStore.LoadBlock(r.Height)
		proof = block.Data.Txs.Proof(int(r.Index))
	}

	return &ctypes.ResultTx{
		Hash:     hash,
		Height:   r.Height,
		Index:    r.Index,
		TxResult: r.Result,
		Tx:       utx,
		Proof:    proof,
	}, nil
}

// 除了查询交易上链情况之外，还查询了交易所在区块的提交时间
func TxTime(ctx *rpctypes.Context, tx []byte, prove bool) (*ctypes.ResultTime, error) {
	utx := types.Tx{
		OriginTx: tx,
	}
	hash := utx.Hash()
	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, fmt.Errorf("transaction indexing is disabled")
	}

	r, err := env.TxIndexer.Get(hash)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	var proof types.TxProof
	if prove {
		block := env.BlockStore.LoadBlock(r.Height)
		proof = block.Data.Txs.Proof(int(r.Index))
	}

	var t int64
	// 每个区块中记录的是上一个区块的提交时间，所以高度为Height的区块提交时间需从Height+1区块中查询
	if heightNow := env.BlockStore.Height(); heightNow < r.Height+1 {
		t = 0
	} else {
		block := env.BlockStore.LoadBlock(r.Height + 1)
		t = block.Time.UnixMicro()
	}

	return &ctypes.ResultTime{
		Hash:     hash,
		Height:   r.Height,
		Index:    r.Index,
		TxResult: r.Result,
		Tx:       r.Tx.OriginTx,
		Proof:    proof,
		Time:     t,
	}, nil
}

// 计算一定范围内区块的吞吐量
func TpsCount(ctx *rpctypes.Context, from, to []byte) (*ctypes.ResultTPS, error) {
	var a, b int64
	var err error
	// 使用strconv解析数字
	// 前端尽量使用strconv将int数字编码为字符串，使用其它方法不确定是否会产生错误
	if a, err = strconv.ParseInt(string(from), 10, 64); err != nil {
		return nil, err
	}
	if b, err = strconv.ParseInt(string(to), 10, 64); err != nil {
		return nil, err
	}
	if a > b {
		return nil, fmt.Errorf("height reverse")
	}
	if env.BlockStore.Height() < b+1 {
		return nil, fmt.Errorf("too large height")
	}
	var num int
	fmt.Println("======================== start query =================================")
	var (
		// commitA：区块a中记录的上一个区块的提交时间；  commitb_1：区块b+1中记录的上一个区块的提交时间
		commitA, commitb_1 time.Time
	)
	for i := a; i <= b+1; i++ {
		block := env.BlockStore.LoadBlock(i)
		if i == a {
			num += len(block.Txs)
			commitA = block.Time
		} else if i == b+1 {
			// 注意，只查询a-b范围内的区块吞吐量，不用计算区块b+1中的交易
			commitb_1 = block.Time
		} else {
			num += len(block.Txs)
			fmt.Println("height : ", i, "size : ", len(block.Txs))
		}
	}
	fmt.Println("======================== end query =================================")

	t := float64(commitb_1.Sub(commitA)) / float64(time.Second)
	tps := float64(num) / t

	return &ctypes.ResultTPS{
		// 吞吐量是浮点数，使用strconv转化为字符串，建议前端使用相同的方法转换为浮点数
		TPS: strconv.FormatFloat(tps, 'e', 30, 64),
	}, nil

}

// 计算一定范围内区块的峰值吞吐量
func TpsmCount(ctx *rpctypes.Context, from, to []byte) (*ctypes.ResultTPS, error) {
	var a, b int64
	var err error
	if a, err = strconv.ParseInt(string(from), 10, 64); err != nil {
		return nil, err
	}
	if b, err = strconv.ParseInt(string(to), 10, 64); err != nil {
		return nil, err
	}
	if a > b {
		return nil, fmt.Errorf("height reverse")
	}
	if env.BlockStore.Height() < b+1 {
		return nil, fmt.Errorf("too large height")
	}
	var (
		lastcommit time.Time
		lastnum    int
		tpsm, tps  float64
	)
	block := env.BlockStore.LoadBlock(a)
	lastcommit = block.Time
	lastnum = len(block.Txs)
	for i := a; i <= b; i++ {
		block = env.BlockStore.LoadBlock(i + 1)
		commitNow := block.Time
		numNow := len(block.Txs)
		dur := commitNow.Sub(lastcommit)
		durSecond := float64(dur / time.Second)
		tps = float64(lastnum) / durSecond
		if tps > tpsm {
			tpsm = tps
		}
		lastcommit, lastnum = commitNow, numNow
	}

	return &ctypes.ResultTPS{
		TPS: strconv.FormatFloat(tpsm, 'e', 30, 64),
	}, nil
}
