package class1

import (
	"context"
	"sync"
)

type SliceQueue[T any] struct {
	mutex    *sync.RWMutex
	data     []T
	size     int
	capacity int
	head     int
	tail     int

	notFull  *Cond
	notEmpty *Cond
	zero     T
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
	m := &sync.RWMutex{}
	return &SliceQueue[T]{
		data:     make([]T, 1, capacity+2),
		capacity: capacity + 2,
		notFull: &Cond{
			Cond: sync.NewCond(m),
		},
		notEmpty: &Cond{
			Cond: sync.NewCond(m),
		},
		mutex: m,
	}
}

//func (s *SliceQueue[T]) Demo() {
//	go func() {
//		// g1
//		ctx1, cancel1 := context.WithTimeout(context.Background(),
//			time.Second)
//		s.Out(ctx1)
//	}()
//
//	go func() {
//		// g3
//		ctx3, cancel3 := context.WithTimeout(context.Background(),
//			time.Second+time.Millisecond*100)
//		s.Out(ctx3)
//	}()
//}

func (s *SliceQueue[T]) Out(ctx context.Context) (T, error) {

	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	s.mutex.Lock()
	for s.isEmpty() {
		err := s.notEmpty.Wait(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}

	var t T
	t = s.data[0]
	s.data[0] = s.zero
	s.data = s.data[1:]
	s.size--

	s.head = (s.head + 1) % s.capacity
	t = s.data[s.head]
	s.data[s.head] = s.zero

	s.notFull.Signal()
	s.mutex.Unlock()
	return t, nil
}

func (s *SliceQueue[T]) EnQueue(ctx context.Context, t T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	s.mutex.Lock()
	for s.IsFull() {
		err := s.notFull.Wait(ctx)
		if err != nil {
			return err
		}
	}

	s.data = append(s.data, t)
	s.size++

	s.data[s.tail] = t
	s.tail = (s.tail + 1) % s.capacity

	s.notEmpty.Signal()
	s.mutex.Unlock()
	return nil
}

func (s *SliceQueue[T]) IsEmpty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isEmpty()
}

func (s *SliceQueue[T]) isEmpty() bool {
	return (s.capacity+s.tail-1)%s.capacity == s.head
}

func (s *SliceQueue[T]) IsFull() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isFull()
}

func (s *SliceQueue[T]) isFull() bool {
	return (s.tail+1)%s.capacity == s.head
}

func (s *SliceQueue[T]) Size() int {
	return s.size
}

func (s *SliceQueue[T]) Close() error {
	//TODO implement me
	panic("implement me")
}
