package queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
)

// SliceQueue 基于切片的实现
type SliceQueue[T any] struct {
	// 在 Java 里面是数组
	data  []T
	head  int
	tail  int
	count int
	zero  T
	mutex *sync.RWMutex
	//notFull  *sync.Cond
	//notEmpty *sync.Cond

	enqueue *semaphore.Weighted
	dequeue *semaphore.Weighted
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
	mutex := &sync.RWMutex{}
	return &SliceQueue[T]{
		data:    make([]T, capacity),
		mutex:   mutex,
		enqueue: semaphore.NewWeighted(int64(capacity)),
		dequeue: semaphore.NewWeighted(int64(capacity)),
	}
}

//	func (q *SliceQueue[T]) Demo() {
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		q.In(ctx, q.zero)
//		cancel()
//	}

func (q *SliceQueue[T]) In(ctx context.Context, v T) error {
	err := q.enqueue.Acquire(ctx, 1)
	if err != nil {
		return err
	}

	// 但凡到了这里，就相当于你已经预留了一个座位
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if ctx.Err() != nil {
		q.enqueue.Release(1)
		return ctx.Err()
	}
	// 没有满才会往下执行
	q.data[q.tail] = v
	q.tail++
	q.count++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
	// 我放了一个元素，我要通知另外一边准备出队的人
	q.dequeue.Release(1)
	//q.notEmpty.Signal()
	return nil
}

// Out ctx 用于超时控制，要么在超时内返回一个数据，要么返回一个 error
func (q *SliceQueue[T]) Out(ctx context.Context) (T, error) {

	err := q.dequeue.Acquire(ctx, 1)
	if err != nil {
		return q.zero, nil
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()
	if ctx.Err() != nil {
		q.dequeue.Release(1)
		return q.zero, ctx.Err()
	}

	front := q.data[q.head]
	q.data[q.head] = q.zero

	q.head++
	q.count--
	if q.head == cap(q.data) {
		q.head = 0
	}
	// 我拿走了一个元素，我就唤醒对面在等待空位的人
	q.enqueue.Release(1)
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
