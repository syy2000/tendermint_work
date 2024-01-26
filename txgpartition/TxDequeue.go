package txgpartition

type TxDequeue struct {
	q *Dequeue
}

func NewTxDequeue() *TxDequeue {
	return &TxDequeue{
		q: NewDequeue(),
	}
}

func (txDequeue *TxDequeue) Size() int {
	return txDequeue.q.Size()
}
func (txDequeue *TxDequeue) Empty() bool {
	return txDequeue.q.Empty()
}
func (txDequeue *TxDequeue) PushBack(n TxNode) {
	txDequeue.q.PushBack(n)
}
func (txDequeue *TxDequeue) PushFront(n TxNode) {
	txDequeue.q.PushFront(n)
}
func (txDequeue *TxDequeue) PopFront() TxNode {
	queueNode := txDequeue.q.PopFront()
	txNode, _ := queueNode.(TxNode)
	return txNode
}
