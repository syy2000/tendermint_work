package types

type TxBlock interface {
	BaseValidate() bool
}

type PoHBlock struct {
	//TODO
}

func (b *PoHBlock) BaseValidate() bool {
	//TODO
	return true
}
