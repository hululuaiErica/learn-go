package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	errFull  = errors.New("队列已经满了")
	errEmpty = errors.New("队列为空")
)

// SliceQueue 先搞一个基于切片的队列实现
// 使用 ring buffer
// 实现阻塞功能
// 到 19:40
type SliceQueue[T any] struct {
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

// 实现阻塞超时功能
// 去洗手间。或者思考这个地方怎么写，
func (s *SliceQueue[T]) Enqueue(ctx context.Context, val T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for (s.head+1)%s.size == s.tail {
		s.writeCond.Wait()
	}
	s.data[s.tail] = val
	s.tail++
	s.readCond.Signal()
	if s.tail >= s.size {
		s.tail = s.tail - s.size
	}
	return nil
}

func (s *SliceQueue[T]) Dequeue(ctx context.Context) (T, error) {
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
func NewSliceQueue[T any](size int) *SliceQueue[T] {
	return &SliceQueue[T]{
		data: make([]T, size),
		size: size,
	}
}

func UseSlice() {
	t := NewSliceQueue[User](10)
	val, err := t.Dequeue(context.Background())
	fmt.Println(val, err)
}
