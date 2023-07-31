package types

type TxTimestamp interface {
	// 返回int64时间戳，数字越小时间越早
	GetTimestamp() int64
}

type PoHTimestamp struct {
	Round   int64
	Input   []byte
	Message []byte
	Out     []byte
}

func (t1 *PoHTimestamp) GetTimestamp() int64 {
	return t1.Round
}

type Seed struct {
	Seed   []byte
	Height int64
	Round  int64
}

type TxWithTimestamp interface {
	GetTx() []byte
	SetTimestamp(t TxTimestamp)
	GetTimestamp() TxTimestamp
}
