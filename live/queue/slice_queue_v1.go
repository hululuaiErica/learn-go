package queue

import (
	"context"
	"github.com/labstack/gommon/log"
)

// SliceQueueV1 先搞一个基于切片的队列实现
// 使用 ring buffer
// 实现阻塞功能
// 到 19:40
type SliceQueueV1[T any] struct {
	// 这边就是我们的数据
	data []T
	head int
	tail int
	size int

	writeCh chan struct{}
	readCh  chan struct{}

	zero T
}

// 实现阻塞超时功能
// 去洗手间。或者思考这个地方怎么写，
// 具体一点，假如说容量是 10
// 1. 连续 Enqueue 十次
func (s *SliceQueueV1[T]) Enqueue(ctx context.Context, val T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.writeCh:
		// 可写
		//defer func() {
		//	s.writeCh <- struct{}{}
		//}()
		s.data[s.tail] = val
		s.tail++
		if s.tail >= s.size {
			s.tail = s.tail - s.size
		}
		// 第二次进来就发不进去了
		// 在这里阻塞了，永久阻塞，而且不会返回超时
		select {
		case s.readCh <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *SliceQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		return s.zero, ctx.Err()
	case <-s.readCh:
		// 可写
		//defer func() {
		//	s.readCh <- struct{}{}
		//}()

		res := s.data[s.head]
		// 你取完之后，你要考虑垃圾回收的问题
		s.data[s.head] = s.zero
		s.head++
		if s.head >= s.size {
			s.head = s.head - s.size
		}
		select {
		case s.writeCh <- struct{}{}:
		default:
			log.Printf("进来嘞这里\n")
			return res, ctx.Err()
		}
		return res, nil
	}
}

//func (s *SliceQueue[T]) Enqueue(ctx context.Context, val T) error {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//	done := false
//	go func() {
//		select {
//		case <-ctx.Done():
//			done = true
//			s.condFull.Broadcast()
//		}
//	}()
//	for (s.w+1)%s.size == s.r {
//		s.condFull.Wait()
//		if done {
//			return errTimeout
//		}
//	}
//	s.buffer[s.w] = val
//	s.w++
//	if s.w >= s.size {
//		s.w = 0
//	}
//	s.condEmpty.Signal()
//	return nil
//}

//go:inline
func NewSliceQueueV1[T any](size int) *SliceQueueV1[T] {
	ch := make(chan struct{}, 10)
	ch <- struct{}{}
	return &SliceQueueV1[T]{
		data:    make([]T, size),
		size:    size,
		writeCh: ch,
		readCh:  make(chan struct{}, 10),
	}
}

//func UseSlice() {
//	t := NewSliceQueueV1[User](10)
//	val, err := t.Dequeue(context.Background())
//	fmt.Println(val, err)
//}
