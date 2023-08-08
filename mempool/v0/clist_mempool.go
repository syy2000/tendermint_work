package v0

import (
	"bytes"
	"errors"
	"sync"
	"sync/atomic"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/clist"
	"github.com/tendermint/tendermint/libs/log"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/tendermint/txgpartition"
)

// CListMempool is an ordered in-memory pool for transactions before they are
// proposed in a consensus round. Transaction validity is checked using the
// CheckTx abci message before the transaction is added to the pool. The
// mempool uses a concurrent list structure for storing transactions that can
// be efficiently accessed by multiple concurrent readers.
type CListMempool struct {
	// Atomic integers
	height   int64 // the last block Update()'d to
	txsBytes int64 // total size of mempool, in bytes
	//modified by syy
	//txsOpNum int64 // 事务中操作的数量

	// notify listeners (ie. consensus) when txs are available
	notifiedTxsAvailable bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty

	config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of
	// CheckTx or ReapMaxBytesMaxGas(ReapMaxTxs) methods.
	updateMtx tmsync.RWMutex
	preCheck  mempool.PreCheckFunc
	postCheck mempool.PostCheckFunc

	txs          *clist.CList // concurrent linked-list of good txs
	proxyAppConn proxy.AppConnMempool

	// Track whether we're rechecking txs.
	// These are not protected by a mutex and are expected to be mutated in
	// serial (ie. by abci responses which are called in serial).
	recheckCursor *clist.CElement // next expected response
	recheckEnd    *clist.CElement // re-checking stops here

	// Map for quick access to txs to record sender in CheckTx.
	// txsMap: txKey -> CElement
	txsMap sync.Map

	// Keep a cache of already-seen txs.
	// This reduces the pressure on the proxyApp.
	cache mempool.TxCache

	logger  log.Logger
	metrics *mempool.Metrics
	//modified by syy
	blockStatusMap sync.Map // 区块状态映射表
	txsConflictMap sync.Map // 事务依赖表
}

// modified by syy
type txsConflictMapValue struct {
	attrValue string
	curTx     []*mempoolTx
	prevTx    []*mempoolTx
	operation string
}

var _ mempool.Mempool = &CListMempool{}

// CListMempoolOption sets an optional parameter on the mempool.
type CListMempoolOption func(*CListMempool)

// NewCListMempool returns a new mempool with the given configuration and
// connection to an application.
func NewCListMempool(
	cfg *config.MempoolConfig,
	proxyAppConn proxy.AppConnMempool,
	height int64,
	options ...CListMempoolOption,
) *CListMempool {

	mp := &CListMempool{
		config:        cfg,
		proxyAppConn:  proxyAppConn,
		txs:           clist.New(),
		height:        height,
		recheckCursor: nil,
		recheckEnd:    nil,
		logger:        log.NewNopLogger(),
		metrics:       mempool.NopMetrics(),
	}

	if cfg.CacheSize > 0 {
		mp.cache = mempool.NewLRUTxCache(cfg.CacheSize)
	} else {
		mp.cache = mempool.NopTxCache{}
	}

	proxyAppConn.SetResponseCallback(mp.globalCb)

	for _, option := range options {
		option(mp)
	}

	return mp
}

// NOTE: not thread safe - should only be called once, on startup
func (mem *CListMempool) EnableTxsAvailable() {
	mem.txsAvailable = make(chan struct{}, 1)
}

// SetLogger sets the Logger.
func (mem *CListMempool) SetLogger(l log.Logger) {
	mem.logger = l
}

// WithPreCheck sets a filter for the mempool to reject a tx if f(tx) returns
// false. This is ran before CheckTx. Only applies to the first created block.
// After that, Update overwrites the existing value.
func WithPreCheck(f mempool.PreCheckFunc) CListMempoolOption {
	return func(mem *CListMempool) { mem.preCheck = f }
}

// WithPostCheck sets a filter for the mempool to reject a tx if f(tx) returns
// false. This is ran after CheckTx. Only applies to the first created block.
// After that, Update overwrites the existing value.
func WithPostCheck(f mempool.PostCheckFunc) CListMempoolOption {
	return func(mem *CListMempool) { mem.postCheck = f }
}

// WithMetrics sets the metrics.
func WithMetrics(metrics *mempool.Metrics) CListMempoolOption {
	return func(mem *CListMempool) { mem.metrics = metrics }
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Lock() {
	mem.updateMtx.Lock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Unlock() {
	mem.updateMtx.Unlock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Size() int {
	return mem.txs.Len()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) SizeBytes() int64 {
	//modified by syy
	return atomic.LoadInt64(&mem.txsBytes)
}

// Lock() must be help by the caller during execution.
func (mem *CListMempool) FlushAppConn() error {
	return mem.proxyAppConn.FlushSync()
}

// XXX: Unsafe! Calling Flush may leave mempool in inconsistent state.
func (mem *CListMempool) Flush() {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	_ = atomic.SwapInt64(&mem.txsBytes, 0)
	mem.cache.Reset()

	for e := mem.txs.Front(); e != nil; e = e.Next() {
		mem.txs.Remove(e)
		e.DetachPrev()
	}

	mem.txsMap.Range(func(key, _ interface{}) bool {
		mem.txsMap.Delete(key)
		return true
	})
}

// TxsFront returns the first transaction in the ordered list for peer
// goroutines to call .NextWait() on.
// FIXME: leaking implementation details!
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsFront() *clist.CElement {
	return mem.txs.Front()
}

// TxsWaitChan returns a channel to wait on transactions. It will be closed
// once the mempool is not empty (ie. the internal `mem.txs` has at least one
// element)
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsWaitChan() <-chan struct{} {
	return mem.txs.WaitChan()
}

// It blocks if we're waiting on Update() or Reap().
// cb: A callback from the CheckTx command.
//
//	It gets called from another goroutine.
//
// CONTRACT: Either cb will get called, or err returned.
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) CheckTx(
	originTx []byte,
	cb func(*abci.Response),
	txInfo mempool.TxInfo,
) error {

	mem.updateMtx.RLock()
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.RUnlock()

	//modified by syy donghao
	//txSize := len(tx)
	//txSize := len(tx.ToProto().OriginTx)
	txSize := len(originTx)
	mmpOriginTx := types.Tx{
		OriginTx: originTx,
		// modified by donghao
		// TODO : fill TxTimehash ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		TxTimehash: nil,
	}
	tx := types.MemTx{
		OriginTx: mmpOriginTx,
		// modified by donghao
		// TODO : fill Other Values ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	}

	if err := mem.isFull(txSize); err != nil {
		return err
	}

	if txSize > mem.config.MaxTxBytes {
		return mempool.ErrTxTooLarge{
			Max:    mem.config.MaxTxBytes,
			Actual: txSize,
		}
	}

	if mem.preCheck != nil {
		if err := mem.preCheck(tx.OriginTx); err != nil {
			return mempool.ErrPreCheck{
				Reason: err,
			}
		}
	}

	// NOTE: proxyAppConn may error if tx buffer is full
	if err := mem.proxyAppConn.Error(); err != nil {
		return err
	}

	if !mem.cache.Push(tx.OriginTx) { // if the transaction already exists in the cache
		// Record a new sender for a tx we've already seen.
		// Note it's possible a tx is still in the cache but no longer in the mempool
		// (eg. after committing a block, txs are removed from mempool but not cache),
		// so we only record the sender for txs still in the mempool.
		if e, ok := mem.txsMap.Load(tx.OriginTx.Key()); ok {
			memTx := e.(*clist.CElement).Value.(*mempoolTx)
			memTx.senders.LoadOrStore(txInfo.SenderID, true)
			// TODO: consider punishing peer for dups,
			// its non-trivial since invalid txs can become valid,
			// but they can spam the same tx with little cost to them atm.
		}
		return mempool.ErrTxInCache
	}

	reqRes := mem.proxyAppConn.CheckTxAsync(abci.RequestCheckTx{Tx: tx.OriginTx.ToProto()})
	reqRes.SetCallback(mem.reqResCb(tx, txInfo.SenderID, txInfo.SenderP2PID, cb))

	return nil
}

// TODO
func (mem *CListMempool) CheckTxReactor(
	rawtx types.Tx,
	cb func(*abci.Response),
	txInfo mempool.TxInfo,
) error {

	mem.updateMtx.RLock()
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.RUnlock()

	//modified by syy donghao
	//txSize := len(tx)
	//txSize := len(tx.ToProto().OriginTx)
	txSize := len(rawtx.OriginTx)
	tx := types.MemTx{
		OriginTx: rawtx,
		// modified by donghao
		// TODO : fill Other Values ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	}

	if err := mem.isFull(txSize); err != nil {
		return err
	}

	if txSize > mem.config.MaxTxBytes {
		return mempool.ErrTxTooLarge{
			Max:    mem.config.MaxTxBytes,
			Actual: txSize,
		}
	}

	if mem.preCheck != nil {
		if err := mem.preCheck(tx.OriginTx); err != nil {
			return mempool.ErrPreCheck{
				Reason: err,
			}
		}
	}

	// NOTE: proxyAppConn may error if tx buffer is full
	if err := mem.proxyAppConn.Error(); err != nil {
		return err
	}

	if !mem.cache.Push(tx.OriginTx) { // if the transaction already exists in the cache
		// Record a new sender for a tx we've already seen.
		// Note it's possible a tx is still in the cache but no longer in the mempool
		// (eg. after committing a block, txs are removed from mempool but not cache),
		// so we only record the sender for txs still in the mempool.
		if e, ok := mem.txsMap.Load(tx.OriginTx.Key()); ok {
			memTx := e.(*clist.CElement).Value.(*mempoolTx)
			memTx.senders.LoadOrStore(txInfo.SenderID, true)
			// TODO: consider punishing peer for dups,
			// its non-trivial since invalid txs can become valid,
			// but they can spam the same tx with little cost to them atm.
		}
		return mempool.ErrTxInCache
	}

	reqRes := mem.proxyAppConn.CheckTxAsync(abci.RequestCheckTx{Tx: tx.OriginTx.ToProto()})
	reqRes.SetCallback(mem.reqResCb(tx, txInfo.SenderID, txInfo.SenderP2PID, cb))

	return nil
}

// Global callback that will be called after every ABCI response.
// Having a single global callback avoids needing to set a callback for each request.
// However, processing the checkTx response requires the peerID (so we can track which txs we heard from who),
// and peerID is not included in the ABCI request, so we have to set request-specific callbacks that
// include this information. If we're not in the midst of a recheck, this function will just return,
// so the request specific callback can do the work.
//
// When rechecking, we don't need the peerID, so the recheck callback happens
// here.
func (mem *CListMempool) globalCb(req *abci.Request, res *abci.Response) {
	if mem.recheckCursor == nil {
		return
	}

	mem.metrics.RecheckTimes.Add(1)
	mem.resCbRecheck(req, res)

	// update metrics
	mem.metrics.Size.Set(float64(mem.Size()))
}

// Request specific callback that should be set on individual reqRes objects
// to incorporate local information when processing the response.
// This allows us to track the peer that sent us this tx, so we can avoid sending it back to them.
// NOTE: alternatively, we could include this information in the ABCI request itself.
//
// External callers of CheckTx, like the RPC, can also pass an externalCb through here that is called
// when all other response processing is complete.
//
// Used in CheckTx to record PeerID who sent us the tx.
func (mem *CListMempool) reqResCb(
	//tx []byte,
	//modified by syy
	tx types.MemTx,
	peerID uint16,
	peerP2PID p2p.ID,
	externalCb func(*abci.Response),
) func(res *abci.Response) {
	return func(res *abci.Response) {
		if mem.recheckCursor != nil {
			// this should never happen
			panic("recheck cursor is not nil in reqResCb")
		}

		mem.resCbFirstTime(tx, peerID, peerP2PID, res)

		// update metrics
		mem.metrics.Size.Set(float64(mem.Size()))

		// passed in by the caller of CheckTx, eg. the RPC
		if externalCb != nil {
			externalCb(res)
		}
	}
}

// Called from:
//   - resCbFirstTime (lock not held) if tx is valid
func (mem *CListMempool) addTx(memTx *mempoolTx) {
	e := mem.txs.PushBack(memTx)
	// modified by syy
	mem.txsMap.Store(memTx.tx.TxId, e)
	atomic.AddInt64(&mem.txsBytes, int64(len(memTx.tx.OriginTx.OriginTx)))
	mem.metrics.TxSizeBytes.Observe(float64(len(memTx.tx.OriginTx.OriginTx)))
}

// Called from:
//   - Update (lock held) if tx was committed
//   - resCbRecheck (lock not held) if tx was invalidated
func (mem *CListMempool) removeTx(tx types.Tx, elem *clist.CElement, removeFromCache bool) {
	mem.txs.Remove(elem)
	elem.DetachPrev()
	mem.txsMap.Delete(tx.Key())
	//modified by syy
	atomic.AddInt64(&mem.txsBytes, int64(-len(tx.OriginTx)))

	if removeFromCache {
		mem.cache.Remove(tx)
	}
}

// RemoveTxByKey removes a transaction from the mempool by its TxKey index.
func (mem *CListMempool) RemoveTxByKey(txKey types.TxKey) error {
	if e, ok := mem.txsMap.Load(txKey); ok {
		memTx := e.(*clist.CElement).Value.(*mempoolTx)
		if memTx != nil {
			mem.removeTx(memTx.tx.OriginTx, e.(*clist.CElement), false)
			return nil
		}
		return errors.New("transaction not found")
	}
	return errors.New("invalid transaction found")
}

func (mem *CListMempool) isFull(txSize int) error {
	var (
		memSize  = mem.Size()
		txsBytes = mem.SizeBytes()
	)

	if memSize >= mem.config.Size || int64(txSize)+txsBytes > mem.config.MaxTxsBytes {
		return mempool.ErrMempoolIsFull{
			NumTxs:      memSize,
			MaxTxs:      mem.config.Size,
			TxsBytes:    txsBytes,
			MaxTxsBytes: mem.config.MaxTxsBytes,
		}
	}

	return nil
}

// callback, which is called after the app checked the tx for the first time.
//
// The case where the app checks the tx for the second and subsequent times is
// handled by the resCbRecheck callback.
func (mem *CListMempool) resCbFirstTime(
	//tx []byte,
	//modified by syy
	tx types.MemTx,
	peerID uint16,
	peerP2PID p2p.ID,
	res *abci.Response,
) {
	switch r := res.Value.(type) {
	case *abci.Response_CheckTx:
		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx.OriginTx, r.CheckTx)
		}
		if (r.CheckTx.Code == abci.CodeTypeOK) && postCheckErr == nil {
			// Check mempool isn't full again to reduce the chance of exceeding the
			// limits.
			//modified by syy
			//Tx转TxNew
			if err := mem.isFull(len(tx.OriginTx.OriginTx)); err != nil {
				// remove from cache (mempool might have a space later)
				mem.cache.Remove(tx.OriginTx)
				mem.logger.Error(err.Error())
				return
			}
			// types.Tx转types.MemTx

			memTx := &mempoolTx{
				height:    mem.height,
				gasWanted: r.CheckTx.GasWanted,
				tx:        tx,
				//modified by syy
				inDegree:  0,
				outDegree: 0,
			}
			//modified by syy
			//查找事务依赖表，找到“对象id+属性”的前序依赖
			for index, txObAndAttr := range tx.TxObAndAttr {
				op := tx.TxOp[index] // 读/写操作
				if v, ok := mem.txsConflictMap.Load(txObAndAttr); ok {
					//arr := strings.Split(v.(string), " ")

					// arr[0]: 属性值 arr[1]: 操作事务 arr[2]: 前序依赖事务 arr[3]: 读写操作
					memTx.isBlock = false
					//memTx.tx.SetTxId(1)
					conflictMapValue := v.(*txsConflictMapValue)
					txArr := conflictMapValue.curTx
					latestTx := txArr[len(txArr)-1] // 最近操作此项的事务
					if conflictMapValue.operation == "read" {
						if op == "read" && latestTx.tx.TxId != tx.TxId { // 直接加入，与现有的读并行
							conflictMapValue.curTx = append(conflictMapValue.curTx, memTx)
							mem.txsConflictMap.Store(txObAndAttr, conflictMapValue)
						} else if op == "write" { // 该操作为写操作，与前面不能并行
							if latestTx.tx.TxId == tx.TxId {
								conflictMapValue.prevTx = conflictMapValue.curTx[:len(conflictMapValue.curTx)-1]
							} else {
								conflictMapValue.prevTx = conflictMapValue.curTx
							}
							//arr[1] = tx.TxId()
							conflictMapValue.curTx = conflictMapValue.curTx[0:0] // 清空
							conflictMapValue.curTx = append(conflictMapValue.curTx, memTx)
							conflictMapValue.operation = "write"
							mem.txsConflictMap.Store(txObAndAttr, conflictMapValue)
						}
					} else if conflictMapValue.operation == "write" { //上一个操作此对象+属性的是事务的写操作，与此事务不能并行
						if latestTx.tx.TxId != tx.TxId { // 若上一个操作此对象+id的还是该事务，不用修改
							conflictMapValue.prevTx = conflictMapValue.curTx
							conflictMapValue.curTx = conflictMapValue.curTx[0:0]
							conflictMapValue.curTx = append(conflictMapValue.curTx, memTx)
							mem.txsConflictMap.Store(txObAndAttr, conflictMapValue)
						}
					}
				} else { // 事务依赖表没有此项，查找区块状态映射表
					if v, ok := mem.blockStatusMap.Load(txObAndAttr); ok {
						//区块状态映射表中的区块号作为前序依赖事务
						blockId := v.(int64)
						prevTx := &mempoolTx{
							isBlock: true,
						}
						prevTx.tx.SetTxId(blockId)
						conflictMapValue := &txsConflictMapValue{
							attrValue: "",
							curTx:     []*mempoolTx{memTx},
							prevTx:    []*mempoolTx{prevTx},
							operation: op,
						}
						mem.txsConflictMap.Store(txObAndAttr, conflictMapValue) // 存入事务依赖表
					}
				}
			}
			//modified by syy
			//生成结点的邻接关系
			for _, txObAndAttr := range tx.TxObAndAttr {
				if v, ok := mem.txsConflictMap.Load(txObAndAttr); ok {
					//arr := strings.Split(v.(string), " ")
					conflictMapValue := v.(*txsConflictMapValue)
					//txArr := conflictMapValue.curTx
					prevTxs := conflictMapValue.prevTx //当前事务在此对象+属性上的前序依赖事务
					for _, prevTx := range prevTxs {
						// if !contains(memTx.conflictTxs, prevTx){
						// 	memTx.conflictTxs = append(memTx.conflictTxs, prevTx)
						// }
						if !containsTx(memTx.parentTxs, prevTx) {
							memTx.parentTxs = append(memTx.parentTxs, prevTx)
							memTx.outDegree += 1
						}
						//根据事务id找到对应的mempoolTx
						if !containsTx(prevTx.childTxs, memTx) {
							prevTx.childTxs = append(prevTx.childTxs, memTx)
							prevTx.inDegree += 1
						}
					}
				}
			}
			memTx.senders.Store(peerID, true)
			//modified by syy
			mem.addTx(memTx)
			mem.logger.Debug(
				"added good transaction",
				"tx", types.Tx(tx.OriginTx).Hash(),
				"res", r,
				"height", memTx.height,
				"total", mem.Size(),
			)
			mem.notifyTxsAvailable()
		} else {
			// ignore bad transaction
			mem.logger.Debug(
				"rejected bad transaction",
				"tx", types.Tx(tx.OriginTx).Hash(),
				"peerID", peerP2PID,
				"res", r,
				"err", postCheckErr,
			)
			mem.metrics.FailedTxs.Add(1)

			if !mem.config.KeepInvalidTxsInCache {
				// remove from cache (it might be good later)
				mem.cache.Remove(tx.OriginTx)
			}
		}

	default:
		// ignore other messages
	}
}

func removeLastElement(s string) {
	panic("unimplemented")
}

// callback, which is called after the app rechecked the tx.
//
// The case where the app checks the tx for the first time is handled by the
// resCbFirstTime callback.
func (mem *CListMempool) resCbRecheck(req *abci.Request, res *abci.Response) {
	switch r := res.Value.(type) {
	case *abci.Response_CheckTx:
		tx := req.GetCheckTx().Tx
		tx1 := types.Tx{
			OriginTx:   tx.OriginTx,
			TxTimehash: types.NewPoHTimestampFromProto(tx.TxTimehash),
		}
		memTx := mem.recheckCursor.Value.(*mempoolTx)

		// Search through the remaining list of tx to recheck for a transaction that matches
		// the one we received from the ABCI application.
		for {
			if bytes.Equal(tx.OriginTx, memTx.tx.OriginTx.OriginTx) {
				// We've found a tx in the recheck list that matches the tx that we
				// received from the ABCI application.
				// Break, and use this transaction for further checks.
				break
			}

			mem.logger.Error(
				"re-CheckTx transaction mismatch",
				//modified by syy
				//"got", types.Tx(tx),
				"got", types.Tx(memTx.tx.OriginTx),
				"expected", memTx.tx,
			)

			if mem.recheckCursor == mem.recheckEnd {
				// we reached the end of the recheckTx list without finding a tx
				// matching the one we received from the ABCI application.
				// Return without processing any tx.
				mem.recheckCursor = nil
				return
			}

			mem.recheckCursor = mem.recheckCursor.Next()
			memTx = mem.recheckCursor.Value.(*mempoolTx)
		}

		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx1, r.CheckTx)
		}

		if (r.CheckTx.Code == abci.CodeTypeOK) && postCheckErr == nil {
			// Good, nothing to do.
		} else {
			// Tx became invalidated due to newly committed block.
			mem.logger.Debug("tx is no longer valid", "tx", types.Tx(tx1).Hash(), "res", r, "err", postCheckErr)
			// NOTE: we remove tx from the cache because it might be good later
			mem.removeTx(tx1, mem.recheckCursor, !mem.config.KeepInvalidTxsInCache)
		}
		if mem.recheckCursor == mem.recheckEnd {
			mem.recheckCursor = nil
		} else {
			mem.recheckCursor = mem.recheckCursor.Next()
		}
		if mem.recheckCursor == nil {
			// Done!
			mem.logger.Debug("done rechecking txs")

			// incase the recheck removed all txs
			if mem.Size() > 0 {
				mem.notifyTxsAvailable()
			}
		}
	default:
		// ignore other messages
	}
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsAvailable() <-chan struct{} {
	return mem.txsAvailable
}

func (mem *CListMempool) notifyTxsAvailable() {
	if mem.Size() == 0 {
		panic("notified txs available but mempool is empty!")
	}
	if mem.txsAvailable != nil && !mem.notifiedTxsAvailable {
		// channel cap is 1, so this will send once
		mem.notifiedTxsAvailable = true
		select {
		case mem.txsAvailable <- struct{}{}:
		default:
		}
	}
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	var (
		totalGas    int64
		runningSize int64
	)

	// TODO: we will get a performance boost if we have a good estimate of avg
	// size per tx, and set the initial capacity based off of that.
	// txs := make([]types.Tx, 0, tmmath.MinInt(mem.txs.Len(), max/mem.avgTxSize))
	txs := make([]types.Tx, 0, mem.txs.Len())
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTx)

		txs = append(txs, memTx.tx.OriginTx)

		dataSize := types.ComputeProtoSizeForTxs([]types.Tx{memTx.tx.OriginTx})

		// Check total size requirement
		if maxBytes > -1 && runningSize+dataSize > maxBytes {
			return txs[:len(txs)-1]
		}

		runningSize += dataSize

		// Check total gas requirement.
		// If maxGas is negative, skip this check.
		// Since newTotalGas < masGas, which
		// must be non-negative, it follows that this won't overflow.
		newTotalGas := totalGas + memTx.gasWanted
		if maxGas > -1 && newTotalGas > maxGas {
			return txs[:len(txs)-1]
		}
		totalGas = newTotalGas
	}
	return txs
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) ReapMaxTxs(max int) types.Txs {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	if max < 0 {
		max = mem.txs.Len()
	}

	txs := make([]types.Tx, 0, tmmath.MinInt(mem.txs.Len(), max))
	for e := mem.txs.Front(); e != nil && len(txs) <= max; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		txs = append(txs, memTx.tx.OriginTx)
	}
	return txs
}

// Lock() must be help by the caller during execution.
func (mem *CListMempool) Update(
	height int64,
	txs types.Txs,
	deliverTxResponses []*abci.ResponseDeliverTx,
	preCheck mempool.PreCheckFunc,
	postCheck mempool.PostCheckFunc,
) error {
	// Set height
	mem.height = height
	mem.notifiedTxsAvailable = false

	if preCheck != nil {
		mem.preCheck = preCheck
	}
	if postCheck != nil {
		mem.postCheck = postCheck
	}

	for i, tx := range txs {
		if deliverTxResponses[i].Code == abci.CodeTypeOK {
			// Add valid committed tx to the cache (if missing).
			_ = mem.cache.Push(tx)
		} else if !mem.config.KeepInvalidTxsInCache {
			// Allow invalid transactions to be resubmitted.
			mem.cache.Remove(tx)
		}

		// Remove committed tx from the mempool.
		//
		// Note an evil proposer can drop valid txs!
		// Mempool before:
		//   100 -> 101 -> 102
		// Block, proposed by an evil proposer:
		//   101 -> 102
		// Mempool after:
		//   100
		// https://github.com/tendermint/tendermint/issues/3322.
		if e, ok := mem.txsMap.Load(tx.Key()); ok {
			mem.removeTx(tx, e.(*clist.CElement), false)
		}
	}

	// Either recheck non-committed txs to see if they became invalid
	// or just notify there're some txs left.
	if mem.Size() > 0 {
		if mem.config.Recheck {
			mem.logger.Debug("recheck txs", "numtxs", mem.Size(), "height", height)
			mem.recheckTxs()
			// At this point, mem.txs are being rechecked.
			// mem.recheckCursor re-scans mem.txs and possibly removes some txs.
			// Before mem.Reap(), we should wait for mem.recheckCursor to be nil.
		} else {
			mem.notifyTxsAvailable()
		}
	}

	// Update metrics
	mem.metrics.Size.Set(float64(mem.Size()))

	return nil
}

func (mem *CListMempool) recheckTxs() {
	if mem.Size() == 0 {
		panic("recheckTxs is called, but the mempool is empty")
	}

	mem.recheckCursor = mem.txs.Front()
	mem.recheckEnd = mem.txs.Back()

	// Push txs to proxyAppConn
	// NOTE: globalCb may be called concurrently.
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		mem.proxyAppConn.CheckTxAsync(abci.RequestCheckTx{
			Tx:   memTx.tx.OriginTx.ToProto(),
			Type: abci.CheckTxType_Recheck,
		})
	}

	mem.proxyAppConn.FlushAsync()
}

// modified by syy
func (mem *CListMempool) blockIdToMemTx(blockId int64) *mempoolTx {
	memTx := &mempoolTx{
		isBlock: true,
	}
	memTx.tx.SetTxId(blockId)
	return memTx
}

//--------------------------------------------------------------------------------

// mempoolTx is a transaction that successfully ran
type mempoolTx struct {
	height    int64       // height that this tx had been validated in
	gasWanted int64       // amount of gas this tx states it will require
	tx        types.MemTx //

	// ids of peers who've sent us this tx (as a map for quick lookups).
	// senders: PeerID -> bool
	senders sync.Map
	//modified by syy
	//conflictTxs []string // 记录的是冲突的事务id
	inDegree  int
	outDegree int
	parentTxs []*mempoolTx
	childTxs  []*mempoolTx
	isBlock   bool
}

// Height returns the height for this transaction
func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}

// modified by syy
func containsTx(arrTx []*mempoolTx, targetTx *mempoolTx) bool {
	for _, item := range arrTx {
		if item == targetTx {
			return true
		}
	}
	return false
}

// modified by donghao
var _ txgpartition.TxNode = (*mempoolTx)(nil)

func (memTx *mempoolTx) ID() int64 {
	return memTx.tx.TxId
}
func (memTx *mempoolTx) Less(other txgpartition.TxNode) bool {
	return memTx.ID() < other.ID()
}
func (memTx *mempoolTx) Equal(other txgpartition.TxNode) bool {
	return memTx.ID() == other.ID()
}
