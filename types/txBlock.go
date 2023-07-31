package types

import (
	"bytes"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/bits"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"io"
)

type TxBlock interface {
	BaseValidate() bool
}

type PoHBlock struct {
	Height        int64
	PoHTimestamps []*PoHTimestamp
	Signature     []byte
}

func (b *PoHBlock) BaseValidate() bool {
	//TODO
	return true
}

func (b *PoHBlock) ToProto() *tmproto.PoHBlock {
	if b == nil {
		return nil
	}
	temp := make([]*tmproto.PoHTimestamp, len(b.PoHTimestamps))
	for i, m := range b.PoHTimestamps {
		temp[i] = m.ToProto()
	}
	return &tmproto.PoHBlock{
		Height:        b.Height,
		PoHTimestamps: temp,
		Signature:     b.Signature,
	}
}

type PoHBlockPart struct {
	Height int64            `json:"height"`
	Total  uint32           `json:"total"`
	Index  uint32           `json:"index"`
	Bytes  tmbytes.HexBytes `json:"bytes"`
	Proof  merkle.Proof     `json:"proof"`
}

type PoHBlockPartSet struct {
	total uint32
	hash  []byte

	mtx           tmsync.Mutex
	parts         []*PoHBlockPart
	partsBitArray *bits.BitArray
	count         uint32
	// a count of the total size (in bytes). Used to ensure that the
	// part set doesn't exceed the maximum block bytes
	byteSize int64
	height   int64
}

func NewPoHBlockPartSetFromData(data []byte, partSize uint32, height int64) *PoHBlockPartSet {
	// divide data into 4kb parts.
	total := (uint32(len(data)) + partSize - 1) / partSize
	parts := make([]*PoHBlockPart, total)
	partsBytes := make([][]byte, total)
	partsBitArray := bits.NewBitArray(int(total))
	for i := uint32(0); i < total; i++ {
		part := &PoHBlockPart{
			Height: height,
			Total:  total,
			Index:  i,
			Bytes:  data[i*partSize : tmmath.MinInt(len(data), int((i+1)*partSize))],
		}
		parts[i] = part
		partsBytes[i] = part.Bytes
		partsBitArray.SetIndex(int(i), true)
	}
	// Compute merkle proofs
	root, proofs := merkle.ProofsFromByteSlices(partsBytes)
	for i := uint32(0); i < total; i++ {
		parts[i].Proof = *proofs[i]
	}
	return &PoHBlockPartSet{
		height:        height,
		total:         total,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         total,
		byteSize:      int64(len(data)),
	}
}

func NewPoHBlockPartSetFromPart(part *PoHBlockPart) *PoHBlockPartSet {
	return &PoHBlockPartSet{
		total:         part.Total,
		height:        part.Height,
		parts:         make([]*PoHBlockPart, part.Total),
		partsBitArray: bits.NewBitArray(int(part.Total)),
		count:         0,
		byteSize:      0,
	}
}

func (ps *PoHBlockPartSet) Count() uint32 {
	if ps == nil {
		return 0
	}
	return ps.count
}

func (ps *PoHBlockPartSet) Total() uint32 {
	if ps == nil {
		return 0
	}
	return ps.total
}

func (ps *PoHBlockPartSet) Hash() []byte {
	if ps == nil {
		return merkle.HashFromByteSlices(nil)
	}
	return ps.hash
}

func (ps *PoHBlockPartSet) AddPart(part *PoHBlockPart) bool {
	if ps == nil {
		return false
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	if part.Height != ps.height {
		return false
	}

	// Invalid part index
	if part.Index >= ps.total {
		return false
	}

	// If part already exists, return false.
	if ps.parts[part.Index] != nil {
		return false
	}

	// // Check hash proof
	// if part.Proof.Verify(ps.Hash(), part.Bytes) != nil {
	// 	return false, &NormalError{}
	// }

	// Add part
	ps.parts[part.Index] = part
	ps.partsBitArray.SetIndex(int(part.Index), true)
	ps.count++
	ps.byteSize += int64(len(part.Bytes))
	return true
}

func (ps *PoHBlockPartSet) GetPart(index int) *PoHBlockPart {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parts[index]
}

func (ps *PoHBlockPartSet) IsComplete() bool {
	return ps.count == ps.total
}

func (ps *PoHBlockPartSet) GetReader() io.Reader {
	if !ps.IsComplete() {
		panic("Cannot GetReader() on incomplete PartSet")
	}
	return NewPoHBlockPartSetReader(ps.parts)
}

type PoHBlockPartSetReader struct {
	i      int
	parts  []*PoHBlockPart
	reader *bytes.Reader
}

func NewPoHBlockPartSetReader(parts []*PoHBlockPart) *PoHBlockPartSetReader {
	return &PoHBlockPartSetReader{
		i:      0,
		parts:  parts,
		reader: bytes.NewReader(parts[0].Bytes),
	}
}

func (psr *PoHBlockPartSetReader) Read(p []byte) (n int, err error) {
	readerLen := psr.reader.Len()
	if readerLen >= len(p) {
		return psr.reader.Read(p)
	} else if readerLen > 0 {
		n1, err := psr.Read(p[:readerLen])
		if err != nil {
			return n1, err
		}
		n2, err := psr.Read(p[readerLen:])
		return n1 + n2, err
	}

	psr.i++
	if psr.i >= len(psr.parts) {
		return 0, io.EOF
	}
	psr.reader = bytes.NewReader(psr.parts[psr.i].Bytes)
	return psr.Read(p)
}
