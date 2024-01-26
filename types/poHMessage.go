package types

type PoHMessage struct {
	Round   int64
	Input   []byte
	Message []byte
	Out     []byte
}
