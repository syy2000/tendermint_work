package types

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
