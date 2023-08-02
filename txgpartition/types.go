package txgpartition

type DequeueNode struct {
	value       interface{}
	front, next *DequeueNode
}

type Dequeue struct {
	root *DequeueNode
	tail *DequeueNode
	size int
}

func NewDequeue() *Dequeue {
	return &Dequeue{
		root: nil,
		tail: nil,
		size: 0,
	}
}

func (q *Dequeue) Size() int {
	return q.size
}

func (q *Dequeue) Empty() bool {
	return q.size == 0
}

func (q *Dequeue) PushBack(v interface{}) {
	vnode := &DequeueNode{
		value: v,
		front: nil,
		next:  nil,
	}
	if q.Empty() {
		q.root = vnode
		q.tail = vnode
	} else {
		q.tail.next, vnode.front = vnode, q.tail
		q.tail = vnode
	}
	q.size++
}

func (q *Dequeue) PushFront(v interface{}) {
	vnode := &DequeueNode{
		value: v,
		front: nil,
		next:  nil,
	}
	if q.Empty() {
		q.root = vnode
		q.tail = vnode
	} else {
		q.root.front, vnode.next = vnode, q.root
		q.root = vnode
	}
	q.size++
}

func (q *Dequeue) PopFront() interface{} {
	if q.Empty() {
		return nil
	} else if q.size == 1 {
		u := q.root
		q.root, q.tail = nil, nil
		q.size = 0
		return u.value
	} else {
		u := q.root
		q.root = u.next
		q.size--
		return u.value
	}
}
