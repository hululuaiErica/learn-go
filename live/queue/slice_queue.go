package queue

import (
	"sync"
)

// SliceQueue 基于切片的实现
type SliceQueue[T any] struct {
	data     []T
	head     int
	tail     int
	count    int
	zero     T
	mutex    *sync.RWMutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
	mutex := &sync.RWMutex{}
	return &SliceQueue[T]{
		data:     make([]T, capacity),
		mutex:    mutex,
		notFull:  sync.NewCond(mutex),
		notEmpty: sync.NewCond(mutex),
	}
}

func (q *SliceQueue[T]) In(v T) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 没有满才会往下执行
	q.data[q.tail] = v
	q.tail++
	q.count++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
	return nil
}

func (q *SliceQueue[T]) Out() (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 不空才会往下执行
	front := q.data[q.head]
	q.data[q.head] = q.zero

	q.head++
	q.count--
	if q.head == cap(q.data) {
		q.head = 0
	}
	return front, nil
}

func (q *SliceQueue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isEmpty()
}

func (q *SliceQueue[T]) isEmpty() bool {
	return q.count == 0
}

func (q *SliceQueue[T]) isFull() bool {
	return q.count == cap(q.data)
}
