package queue

import (
	"context"
	"sync"
	"time"
)

type DelayQueueV4[R any, T DelayableV4[R]] struct {
	priorityQueue *PriorityQueue[T]
	enqueueChan   chan T
	dequeueChan   chan T
	transferChan  chan T
	closeSignal   chan struct{}
	closeWG       sync.WaitGroup
	zero          T
}

type DelayableV4[T any] interface {
	// Deadline 是还剩下多少过期时间
	// 还要延迟多久
	Deadline() time.Time
	Compare(DelayableV4[T]) int
	Value() T
}

var _ DelayableV4[int] = &ElementV4[int]{}

type ElementV4[T any] struct {
	value    T
	deadline time.Time
}

func (elm *ElementV4[T]) Deadline() time.Time {
	return elm.deadline
}

func (elm *ElementV4[T]) Value() T {
	return elm.value
}

func (elm *ElementV4[T]) Compare(dst DelayableV4[T]) int {
	if elm.deadline.Before(dst.Deadline()) {
		return -1
	} else if elm.deadline.After(dst.Deadline()) {
		return 1
	} else {
		return 0
	}
}

func NewDelayQueueV4[R any, T DelayableV4[R]](cap int) *DelayQueueV4[R, T] {
	queue := &DelayQueueV4[R, T]{
		enqueueChan:  make(chan T, 1),
		dequeueChan:  make(chan T, 1),
		transferChan: make(chan T, 1),
		closeSignal:  make(chan struct{}),
		priorityQueue: NewPriorityQueue[T](cap, func(src T, dst T) int {
			return src.Compare(dst)
		}),
	}
	go queue.Run()
	return queue
}

func (s *DelayQueueV4[R, T]) Close() {
	close(s.closeSignal)
	s.closeWG.Wait()
}

func (s *DelayQueueV4[R, T]) Run() {
	const minTimerDelay = time.Millisecond
	var delay time.Duration
	s.closeWG.Add(1)
	defer s.closeWG.Done()

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		peekedElm, err := s.priorityQueue.Peek()
		if err != nil {
			delay = 0
		} else {
			delay = peekedElm.Deadline().Sub(time.Now())
		}
		if delay > minTimerDelay {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(delay)
			if s.priorityQueue.isFull() {
				select {
				case <-s.closeSignal:
					return
				case <-timer.C:
				}
			} else {
				select {
				case enqueuedElm := <-s.enqueueChan:
					_ = s.priorityQueue.Enqueue(enqueuedElm)
					continue
				case <-s.closeSignal:
					return
				case <-timer.C:
				}
			}
		}
		switch {
		case s.priorityQueue.isFull():
			select {
			case s.dequeueChan <- peekedElm:
				_, _ = s.priorityQueue.Dequeue()
			case <-s.closeSignal:
				return
			}
		case s.priorityQueue.isEmpty():
			select {
			case enqueuedElm := <-s.enqueueChan:
				_ = s.priorityQueue.Enqueue(enqueuedElm)
			case <-s.closeSignal:
				return
			}
		default:
			select {
			case s.dequeueChan <- peekedElm:
				_, _ = s.priorityQueue.Dequeue()
			case enqueuedElm := <-s.enqueueChan:
				_ = s.priorityQueue.Enqueue(enqueuedElm)
			case <-s.closeSignal:
				return
			}
		}
	}
}

func (s *DelayQueueV4[R, T]) Dequeue(ctx context.Context) (T, error) {
	select {
	case elm := <-s.dequeueChan:
		return elm, nil
	case elm := <-s.transferChan:
		return elm, nil
	case <-ctx.Done():
		return s.zero, ErrEmptyQueue
	}
}

func (s *DelayQueueV4[R, T]) Enqueue(ctx context.Context, val T) error {
	delay := val.Deadline().Sub(time.Now())
	if delay > 0 {
		timer := time.NewTimer(delay)
		select {
		case s.enqueueChan <- val:
			return nil
		case <-ctx.Done():
			return ErrOutOfCapacity
		case <-timer.C:
		}
	}
	select {
	case s.enqueueChan <- val:
		return nil
	case s.transferChan <- val:
		return nil
	case <-ctx.Done():
		return ErrOutOfCapacity
	}
}
