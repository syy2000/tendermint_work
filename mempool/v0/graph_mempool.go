package v0

/*
last modified : 2023/8/11 AM
name : donghao
*/
import (
	txp "github.com/tendermint/tendermint/txgpartition"
)

var _ txp.TxGraph = (*CListMempool)(nil)

func MustMempoolTx(n txp.TxNode) *mempoolTx {
	if o, ok := n.(*mempoolTx); ok {
		return o
	} else if n == nil {
		return nil
	} else {
		panic("this should not happen! a non mempoolTx is delivered by txgpartition!")
	}
}

func (mmp *CListMempool) IsBlockNode(n txp.TxNode) bool {
	return MustMempoolTx(n).isBlock
}
