package queue

import (
	"context"
	"sync"
)

// SliceQueueV2 先搞一个基于切片的队列实现
// 使用 ring buffer
// 实现阻塞功能
// 到 19:40
type SliceQueueV2[T any] struct {
	// 这边就是我们的数据
	data []T
	head int
	tail int
	size int

	mutex     *sync.Mutex
	readCond  *sync.Cond
	writeCond *sync.Cond

	zero T
}

func (s *SliceQueueV2[T]) Enqueue(ctx context.Context, val T) error {
	timeoutCh := make(chan struct{}, 1)
	ch := s.enqueue(timeoutCh, val)
	select {
	case <-ctx.Done():
		timeoutCh <- struct{}{}
		// 你在这里返回的时候。还是有可能，入队成功了
		return ctx.Err()
	case <-ch:
		return nil
	}
}

// 实现阻塞超时功能
// 去洗手间。或者思考这个地方怎么写，
func (s *SliceQueueV2[T]) enqueue(timeoutCh chan struct{}, val T) chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		for (s.head+1)%s.size == s.tail {
			s.writeCond.Wait()
		}
		select {
		case <-timeoutCh:
			return
		default:
		}
		s.data[s.tail] = val
		s.tail++
		s.readCond.Signal()
		if s.tail >= s.size {
			s.tail = s.tail - s.size
		}
		ch <- struct{}{}
	}()
	return ch
}

func (s *SliceQueueV2[T]) Dequeue(ctx context.Context) (T, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for s.tail == s.head {
		s.readCond.Wait()
	}
	res := s.data[s.head]
	// 你取完之后，你要考虑垃圾回收的问题
	s.data[s.head] = s.zero
	s.head++
	s.writeCond.Signal()
	if s.head >= s.size {
		s.head = s.head - s.size
	}
	return res, nil
}

//go:inline
func NewSliceQueueV2[T any](size int) *SliceQueueV2[T] {
	return &SliceQueueV2[T]{
		data: make([]T, size),
		size: size,
	}
}
