package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/json"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// TxKeySize is the size of the transaction key index
const TxKeySize = sha256.Size

type (
	// Tx is an arbitrary byte array.
	// NOTE: Tx has no types at this level, so when wire encoded it's just length-prefixed.
	// Might we want types here ?
	//modified by syy
	Tx struct {
		OriginTx   []byte
		TxTimehash *PoHTimestamp
	}
	Value struct {
		Cid         string
		DataFeature string
		DataValue   string
	}
	//需要由Tx拆解转化而来，再进而转为MempoolTx
	MemTx struct {
		TxId        int64    // primary key
		TxOp        []string // operation: read write
		TxObAndAttr []string // dataobject: id-attr
		TxValue     []Value
		TxTimehash  *PoHTimestamp
		OriginTx    Tx
	}

	// TxKey is the fixed length array key used as an index.
	TxKey [TxKeySize]byte
)

func (tx MemTx) SetTxId(BlockId int64) {
	tx.TxId = BlockId
}

func (memTx MemTx) UnmarshalJSON(data []byte) error {
	tmpTx := struct {
		TxId        int64    `json:"txid"`
		TxOp        []string `json:"txop"`
		TxObAndAttr []string `json:"txobandattr"`
		TxValue     []Value  `json:"txvalue"` // must be map[string]interface{} or []interface{}
	}{}
	err := json.Unmarshal(data, &tmpTx)
	if err != nil {
		return err
	}
	memTx.TxId = tmpTx.TxId
	memTx.TxOp = tmpTx.TxOp
	memTx.TxObAndAttr = tmpTx.TxObAndAttr
	memTx.TxValue = tmpTx.TxValue
	return nil
}

// Tx转MemTx
func txToMemTx(tx Tx) MemTx {
	memTx := &MemTx{
		TxTimehash: tx.TxTimehash,
		OriginTx:   tx,
	}
	memTx.UnmarshalJSON(tx.OriginTx)
	return *memTx
}

// func (txValue Value) ToProto() *tmproto.Tx_Value {
// 	tp := new(tmproto.Tx_Value)
// 	tp.Cid = txValue.Cid
// 	tp.DataFeature = txValue.DataFeature
// 	tp.DataValue = txValue.DataValue
// 	return tp
// }

// // ToProto converts Data to protobuf
// func (tx Tx) ToProto() tmproto.Tx {
// 	tp := new(tmproto.Tx)

// 	tp.TxId = tx.TxId
// 	if len(tx.TxOp) > 0 {
// 		txBzs := make([]string, len(tx.TxOp))
// 		for i := range tx.TxOp {
// 			txBzs[i] = tx.TxOp[i]
// 		}
// 		tp.TxOp = txBzs
// 	}
// 	if len(tx.TxObAndAttr) > 0 {
// 		txBzs := make([]string, len(tx.TxObAndAttr))
// 		for i := range tx.TxObAndAttr {
// 			txBzs[i] = tx.TxObAndAttr[i]
// 		}
// 		tp.TxOp = txBzs
// 	}
// 	if len(tx.TxValue) > 0 {
// 		txBzs := make([]*tmproto.Tx_Value, len(tx.TxValue))
// 		for i := range tx.TxValue {
// 			txBzs[i] = tx.TxValue[i].ToProto()
// 		}
// 		tp.TxValue = txBzs
// 	}
// 	tp.TxTimehash = (*tmproto.PoHTimestamp)(tx.TxTimehash)
// 	return *tp
// }

func (tx Tx) ToProto() *tmproto.Tx {
	tp := new(tmproto.Tx)
	tp.OriginTx = tx.OriginTx
	tp.TxTimehash = (*tmproto.PoHTimestamp)(tx.TxTimehash)
	return tp
}

func NewTxFromProto(protoTx *tmproto.Tx) *Tx {
	if protoTx == nil {
		return nil
	}
	return &Tx{
		OriginTx:   protoTx.OriginTx,
		TxTimehash: (*PoHTimestamp)(protoTx.TxTimehash),
	}
}

// Hash computes the TMHASH hash of the wire encoded transaction.
func (tx Tx) Hash() []byte {
	return tmhash.Sum(tx.OriginTx)
}

func (tx Tx) Key() TxKey {
	return sha256.Sum256(tx.OriginTx)
}

// String returns the hex-encoded transaction as a string.
// modified by syy
func (tx Tx) String() string {
	return fmt.Sprintf("Tx{%X}", []byte(tx.OriginTx))
}

// modified by syy
// func (tx Tx) SetTxId(blockId int64) {
// 	tx.TxId = blockId
// }

// Txs is a slice of Tx.
type Txs []Tx

// Hash returns the Merkle root hash of the transaction hashes.
// i.e. the leaves of the tree are the hashes of the txs.
func (txs Txs) Hash() []byte {
	// These allocations will be removed once Txs is switched to [][]byte,
	// ref #2603. This is because golang does not allow type casting slices without unsafe
	txBzs := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		txBzs[i] = txs[i].Hash()
	}
	return merkle.HashFromByteSlices(txBzs)
}

// Index returns the index of this transaction in the list, or -1 if not found
func (txs Txs) Index(tx Tx) int {
	for i := range txs {
		//modified by syy
		if bytes.Equal(txs[i].OriginTx, tx.OriginTx) {
			//if Equal(txs[i], tx) {
			return i
		}
	}
	return -1
}

// IndexByHash returns the index of this transaction hash in the list, or -1 if not found
func (txs Txs) IndexByHash(hash []byte) int {
	for i := range txs {
		if bytes.Equal(txs[i].Hash(), hash) {
			return i
		}
	}
	return -1
}

// Proof returns a simple merkle proof for this node.
// Panics if i < 0 or i >= len(txs)
// TODO: optimize this!
func (txs Txs) Proof(i int) TxProof {
	l := len(txs)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bzs[i] = txs[i].Hash()
	}
	root, proofs := merkle.ProofsFromByteSlices(bzs)

	return TxProof{
		RootHash: root,
		Data:     txs[i],
		Proof:    *proofs[i],
	}
}

// TxProof represents a Merkle proof of the presence of a transaction in the Merkle tree.
type TxProof struct {
	RootHash tmbytes.HexBytes `json:"root_hash"`
	Data     Tx               `json:"data"`
	Proof    merkle.Proof     `json:"proof"`
}

// Leaf returns the hash(tx), which is the leaf in the merkle tree which this proof refers to.
func (tp TxProof) Leaf() []byte {
	return tp.Data.Hash()
}

// Validate verifies the proof. It returns nil if the RootHash matches the dataHash argument,
// and if the proof is internally consistent. Otherwise, it returns a sensible error.
func (tp TxProof) Validate(dataHash []byte) error {
	if !bytes.Equal(dataHash, tp.RootHash) {
		return errors.New("proof matches different data hash")
	}
	if tp.Proof.Index < 0 {
		return errors.New("proof index cannot be negative")
	}
	if tp.Proof.Total <= 0 {
		return errors.New("proof total must be positive")
	}
	valid := tp.Proof.Verify(tp.RootHash, tp.Leaf())
	if valid != nil {
		return errors.New("proof is not internally consistent")
	}
	return nil
}

func (tp TxProof) ToProto() tmproto.TxProof {

	pbProof := tp.Proof.ToProto()

	pbtp := tmproto.TxProof{
		RootHash: tp.RootHash,
		//Data:     tp.Data,
		//modified by syy
		Data:  tp.Data.ToProto(),
		Proof: pbProof,
	}

	return pbtp
}
func TxProofFromProto(pb tmproto.TxProof) (TxProof, error) {

	pbProof, err := merkle.ProofFromProto(pb.Proof)
	if err != nil {
		return TxProof{}, err
	}
	//modified by syy
	tx := Tx{
		OriginTx:   pb.Data.OriginTx,
		TxTimehash: (*PoHTimestamp)(pb.Data.TxTimehash),
	}
	pbtp := TxProof{
		RootHash: pb.RootHash,
		Data:     tx,
		Proof:    *pbProof,
	}

	return pbtp, nil
}

// ComputeProtoSizeForTxs wraps the transactions in tmproto.Data{} and calculates the size.
// https://developers.google.com/protocol-buffers/docs/encoding
func ComputeProtoSizeForTxs(txs []Tx) int64 {
	data := Data{Txs: txs}
	pdData := data.ToProto()
	return int64(pdData.Size())
}

// modified by syy
// func Equal(a, b Tx) bool {
// 	if a.TxId == b.TxId {
// 		return true
// 	}
// 	return false
// }
