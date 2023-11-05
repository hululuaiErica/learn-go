package queue

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrUnknown = errors.New("unknown error")
)

const minTimerDelay = 1 * time.Millisecond
const maxTimerDelay = time.Nanosecond * math.MaxInt64

func (p *PriorityQueue[T]) IsFull() bool {
	return p.isFull()
}

func (p *PriorityQueue[T]) IsEmpty() bool {
	return p.isEmpty()
}

type DelayQueueV3[R any, T DelayableV3[R]] struct {
	enqueuedSignal chan struct{}
	dequeuedSignal chan struct{}
	transferChan   chan T
	dequeueMutex   sync.Mutex
	globalMutex    sync.Mutex
	priorityQueue  *PriorityQueue[T]
	zero           T
}

type DelayableV3[T any] interface {
	// Deadline 是还剩下多少过期时间
	// 还要延迟多久
	Deadline() time.Time
	Compare(DelayableV3[T]) int
	Value() T
}

var _ DelayableV3[int] = &ElementV3[int]{}

type ElementV3[T any] struct {
	value    T
	deadline time.Time
}

func (elm *ElementV3[T]) Deadline() time.Time {
	return elm.deadline
}

func (elm *ElementV3[T]) Value() T {
	return elm.value
}

func (elm *ElementV3[T]) Compare(dst DelayableV3[T]) int {
	if elm.deadline.Before(dst.Deadline()) {
		return -1
	} else if elm.deadline.After(dst.Deadline()) {
		return 1
	} else {
		return 0
	}
}

func NewDelayQueueV3[R any, T DelayableV3[R]](cap int) *DelayQueueV3[R, T] {
	return &DelayQueueV3[R, T]{
		enqueuedSignal: make(chan struct{}),
		dequeuedSignal: make(chan struct{}),
		transferChan:   make(chan T),
		priorityQueue: NewPriorityQueue[T](cap, func(src T, dst T) int {
			return src.Compare(dst)
		}),
	}
}

func (s *DelayQueueV3[R, T]) Dequeue(ctx context.Context) (T, error) {
	// 出队锁：避免因重复获取队列头部同一元素降低性能
	s.dequeueMutex.Lock()
	defer s.dequeueMutex.Unlock()

	for {
		// 全局锁：避免入队和出队信号的重置与激活出现并发问题
		s.globalMutex.Lock()
		if ctx.Err() != nil {
			s.globalMutex.Unlock()
			return s.zero, ErrEmptyQueue
		}

		// 接收直接转发的不需要延迟的新元素
		select {
		case elm := <-s.transferChan:
			s.globalMutex.Unlock()
			return elm, nil
		default:
		}

		// 延迟时间初始值为 maxTimerDelay, 表示队列为空
		delay := maxTimerDelay
		switch elm, err := s.priorityQueue.Peek(); err {
		case nil:
			now := time.Now()
			delay = elm.Deadline().Sub(now)
			if delay < minTimerDelay {
				// 满足延迟条件，头部元素出队后直接返回
				_, _ = s.dequeue()
				s.globalMutex.Unlock()
				return elm, err
			}
		case ErrEmptyQueue:
		default:
			s.globalMutex.Unlock()
			return s.zero, err
		}
		// 重置入队信号，避免历史信号干扰
		select {
		case <-s.enqueuedSignal:
		default:
		}
		s.globalMutex.Unlock()

		if delay == maxTimerDelay {
			// 队列为空, 等待新元素
			select {
			case elm := <-s.transferChan:
				return elm, nil
			case <-s.enqueuedSignal:
				continue
			case <-ctx.Done():
				return s.zero, ErrEmptyQueue
			}
		} else if delay >= minTimerDelay {
			// 等待时间到期或新元素加入
			timer := time.NewTimer(delay)
			select {
			case elm := <-s.transferChan:
				return elm, nil
			case <-s.enqueuedSignal:
				continue
			case <-timer.C:
				continue
			case <-ctx.Done():
				return s.zero, ErrEmptyQueue
			}
		} else {
			panic(ErrUnknown)
		}
	}
}

func (s *DelayQueueV3[R, T]) dequeue() (T, error) {
	elm, err := s.priorityQueue.Dequeue()
	if err != nil {
		return s.zero, err
	}
	select {
	case s.dequeuedSignal <- struct{}{}:
	default:
	}
	return elm, nil
}

func (s *DelayQueueV3[R, T]) enqueue(val T) error {
	if err := s.priorityQueue.Enqueue(val); err != nil {
		return err
	}
	select {
	case s.enqueuedSignal <- struct{}{}:
	default:
	}
	return nil
}

func (s *DelayQueueV3[R, T]) Enqueue(ctx context.Context, val T) error {
	for {
		// 全局锁：避免入队和出队信号的重置与激活出现并发问题
		s.globalMutex.Lock()

		if ctx.Err() != nil {
			s.globalMutex.Unlock()
			return ErrOutOfCapacity
		}

		// 如果队列未满，入队后直接返回
		if !s.priorityQueue.IsFull() {
			err := s.enqueue(val)
			s.globalMutex.Unlock()
			return err
		}
		// 队列已满，重置出队信号，避免受到历史信号影响
		select {
		case <-s.dequeuedSignal:
		default:
		}
		s.globalMutex.Unlock()

		if delay := val.Deadline().Sub(time.Now()); delay >= minTimerDelay {
			// 新元素需要延迟，等待退出信号、出队信号和到期信号
			timer := time.NewTimer(delay)
			select {
			case <-s.dequeuedSignal:
				// 收到出队信号，从头开始尝试入队
				continue
			case <-timer.C:
				// 新元素不再需要延迟
			case <-ctx.Done():
				return ErrOutOfCapacity
			}
		} else {
			// 新元素不需要延迟，等待转发成功信号、出队信号和退出信号
			select {
			case s.transferChan <- val:
				// 新元素转发成功，直接返回（避免队列满且元素未到期导致新元素长时间无法入队）
				return nil
			case <-s.dequeuedSignal:
				// 收到出队信号，从头开始尝试入队
				continue
			case <-ctx.Done():
				return ErrOutOfCapacity
			}
		}
	}
}
