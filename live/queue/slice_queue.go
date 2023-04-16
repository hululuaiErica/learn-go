package queue

import (
	"context"
	"sync"
	"time"
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

func (q *SliceQueue[T]) Demo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	q.In(ctx, q.zero)
	cancel()
}
func (q *SliceQueue[T]) In(ctx context.Context, v T) error {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	for q.isFull() {
		// 如果是队列是满的，我就在这里等着
		// 有人出队了，我就会被唤醒
		q.notFull.Wait()
	}
	// 没有满才会往下执行
	q.data[q.tail] = v
	q.tail++
	q.count++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
	// 我放了一个元素，我要通知另外一边准备出队的人
	q.notEmpty.Signal()
	return nil
}

func (q *SliceQueue[T]) Out(ctx context.Context) (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for q.isEmpty() {
		// 如果队列是 empty，我就在这儿等着
		// 有人入队了，我就会被唤醒
		q.notEmpty.Wait()

		// 我被唤醒的时候，队列里面是不是一定有元素
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
	q.notFull.Signal()
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
