package types

type TxTimestamp interface {
	// 小于返回-1，等于返回0，大于返回1
	Compare(t2 *TxTimestamp) int
	// 返回int64时间戳，数字越小时间越早
	GetTimestamp() int64
}

type PoHTimestamp struct {
	Round   int64
	Input   []byte
	Message []byte
	Out     []byte
}

func (t1 *PoHTimestamp) Compare(t2 *PoHTimestamp) int {
	if t1.Round < t2.Round {
		return -1
	} else if t1.Round == t2.Round {
		return 0
	}
	return 1
}

func (t1 *PoHTimestamp) GetTimestamp() int64 {
	return t1.Round
}

type Seed struct {
	seed []byte
}
