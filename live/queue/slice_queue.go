package queue

// SliceQueue 基于切片的实现
type SliceQueue[T any] struct {
	data []T
	head int
	tail int
	count int
	zero T
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T]{
	return &SliceQueue[T]{
		data: make([]T, capacity),
	}
}

func (q *SliceQueue[T]) In(v T) {
	if q.count == cap(q.data) {
		panic("满了")
	}
	q.data[q.tail] = v
	q.tail++
	q.count ++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
}

func (q *SliceQueue[T]) Out() T {
	if q.IsEmpty() {
		panic("queue is empty")
	}
	front := q.data[q.head]

	q.data[q.head] = q.zero

	q.head ++
	q.count --
	if q.head == cap(q.data) {
		q.head = 0
	}
	return front
}

func (q *SliceQueue[T]) IsEmpty() bool {
	return q.count == 0
}
