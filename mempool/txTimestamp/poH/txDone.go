package poH

import (
	"bytes"
	"fmt"
	// "fmt"

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
	// fmt.Println("-------------------------------   add %v   --------------------------------", tx.GetTimestamp().GetTimestamp())
	v, ok := t.txMap.Get(tx)
	if !ok {
		v = 1
	} else {
		v = v.(int) + 1
	}
	t.txMap.Put(tx, v)
}

func (t *TxDone) Done(tx types.TxWithTimestamp) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	// fmt.Println("-------------------------------   done %v  --------------------------------", tx.GetTimestamp().GetTimestamp())
	v, ok := t.txMap.Get(tx)
	if !ok {
		fmt.Printf("------------------------------ error! 删除不存在的key time=%v tx=%v ----------------------------\n", tx.GetTimestamp().GetTimestamp(), string(tx.GetTx()))
	}
	value := v.(int)
	t.txMap.Remove(tx)
	if value > 1 {
		value = value - 1
		t.txMap.Put(tx, v)
	}
}

func (t *TxDone) GetNowMin() (int64, bool) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	txkey, _ := t.txMap.Min()
	if txkey == nil {
		return 0, false
	}
	tx := txkey.(types.TxWithTimestamp)
	return tx.GetTimestamp().GetTimestamp(), true
}
