package poH

import (
	"bytes"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
)

type TxDone struct {
	txMap *treemap.Map
	mtx   sync.RWMutex
}

func txWithTimestampComparator(a, b interface{}) int {
	txt1 := a.(types.TxWithTimestamp)
	txt2 := b.(types.TxWithTimestamp)
	t1 := txt1.GetTimestamp().GetTimestamp()
	t2 := txt2.GetTimestamp().GetTimestamp()
	if t1 < t2 {
		return -1
	}
	if t1 > t2 {
		return 1
	}
	tx1 := txt1.GetTx()
	tx2 := txt2.GetTx()
	return bytes.Compare(tx1, tx2)
}

func NewTxDone() *TxDone {
	txMap := treemap.NewWith(txWithTimestampComparator)
	return &TxDone{
		txMap: txMap,
	}
}

func (t *TxDone) AddTxToTxDone(tx types.TxWithTimestamp) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.txMap.Put(tx, 1)
}

func (t *TxDone) Done(tx types.TxWithTimestamp) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.txMap.Remove(tx)
}

func (t *TxDone) GetNowMin() (int64, bool) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return 0, true
}
