package queue

import (
	"context"
	"sync"
)

// SliceQueue 基于切片的实现
type SliceQueue[T any] struct {
	data  []T
	head  int
	tail  int
	count int
	zero  T
	mutex *sync.RWMutex
	cond  *sync.Cond
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
	mutex := &sync.RWMutex{}
	return &SliceQueue[T]{
		data:  make([]T, capacity),
		mutex: mutex,
		cond:  sync.NewCond(mutex),
	}
}

//	func (q *SliceQueue[T]) Demo() {
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		q.In(ctx, q.zero)
//		cancel()
//	}
func (q *SliceQueue[T]) In(ctx context.Context, v T) error {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	for q.isFull() {
		q.cond.Wait()
	}
	// 没有满才会往下执行
	q.data[q.tail] = v
	q.tail++
	q.count++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
	// 我放了一个元素，我要通知另外一边准备出队的人
	q.cond.Signal()
	return nil
}

// Out ctx 用于超时控制，要么在超时内返回一个数据，要么返回一个 error
func (q *SliceQueue[T]) Out(ctx context.Context) (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for q.isEmpty() {
		q.cond.Wait()
	}

	// 不空才会往下执行
	front := q.data[q.head]
	q.data[q.head] = q.zero

	q.head++
	q.count--
	if q.head == cap(q.data) {
		q.head = 0
	}
	// 我拿走了一个元素，我就唤醒对面在等待空位的人
	q.cond.Signal()
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
