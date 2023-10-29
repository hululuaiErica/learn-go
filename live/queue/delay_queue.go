package queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"
)

type DelayQueueV1[T DelayableV1] struct {
	q          *PriorityQueue[T]
	lock       sync.RWMutex
	enqueueCap *semaphore.Weighted
	dequeueCap *semaphore.Weighted
	zero       T
}

// Enqueue 先来实现
func (d *DelayQueueV1[T]) Enqueue(ctx context.Context, val T) error {
	err := d.enqueueCap.Acquire(ctx, 1)
	if err != nil {
		// 这边就是满了
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if ctx.Err() != nil {
		d.enqueueCap.Release(1)
		return ctx.Err()
	}

	err = d.q.Enqueue(val)
	// 基本不可能的分支
	if err != nil {
		return err
	}
	d.dequeueCap.Release(1)
	return nil
}

// Dequeue 这边该怎么写？
// 或者说，你觉得步骤应该怎么样？
// 下面这个实现还没有考虑过期时间
// 算法：
// 1. 先检测队列有没有元素，没有要阻塞，直到超时，或者拿到元素
// 2. 有元素，你是不是要看一眼，队头的元素的过期时间有没有到
// 2.1 如果过期时间到了，直接出队并且返回
// 2.2 如果过期时间没到，阻塞直到过期时间到了
// 2.2.1 如果在等待的时候，有新元素到了，就要看一眼新元素的过期时间是不是更短
// 2.2.2 如果等待的时候，ctx 超时了，那么就直接返回超时错误

// 到 20:20 分，大家写一下
func (d *DelayQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	err := d.dequeueCap.Acquire(ctx, 1)
	if err != nil {
		return d.zero, err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if ctx.Err() != nil {
		return d.zero, ctx.Err()
	}
	// 看一眼队头的元素，但是没有出队。
	d.q.Peek()
	val, err := d.q.Dequeue()
	// 基本不可能的分支
	if err != nil {
		return d.zero, err
	}
	d.enqueueCap.Release(1)
	return val, nil
}

func NewDelayQueueV1[T DelayableV1](c int) *DelayQueueV1[T] {
	return &DelayQueueV1[T]{
		q: NewPriorityQueue[T](c, func(src T, dst T) int {
			// 这个地方怎么搞？
			// 比较元素优先级
			// 拿到 src 的延迟时间
			// 拿到 dst 的延迟时间
			//getDelay(src)
			srcDelay := src.Deadline()
			dstDelay := dst.Deadline()
			//now := time.Now()
			// 你要仔细处理，会不会 Sub 得到负数的问题
			if srcDelay.Before(dstDelay) {
				return -1
			} else if srcDelay.After(dstDelay) {
				return 1
			}
			return 0
		}),
	}
}

type DelayableV1 interface {
	// Deadline 是还剩下多少过期时间
	// 还要延迟多久
	Deadline() time.Time
}

type Delayable interface {
	// Delay 是还剩下多少过期时间
	// 还要延迟多久
	Delay() time.Duration
}

// DelayQueue 出队的时候，只有到了过期时间的才会出队，
// 没到过期时间的就不会出队
// 基本原理很简单：
// 利用延迟时间作为优先级，然后实现 Dequeue 只在过期之后，弹出元素
type DelayQueue[T Delayable] struct {
	q *PriorityQueue[T]
}

func NewDelayQueue[T Delayable](c int) *DelayQueue[T] {
	return &DelayQueue[T]{
		q: NewPriorityQueue[T](c, func(src T, dst T) int {
			// 这个地方怎么搞？
			// 比较元素优先级
			// 拿到 src 的延迟时间
			// 拿到 dst 的延迟时间
			//getDelay(src)
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			return int(srcDelay - dstDelay)
		}),
	}
}

type elem struct {
	ddl time.Time
}

func (e *elem) Delay() time.Duration {
	now := time.Now()
	return e.ddl.Sub(now)
}

//type Element[T any] struct {
//	Data    T
//	Timeout time.Duration
//}

// timeout 是一分钟，但是我是 01:00 加入进去队列的
// timeout 是十秒钟，但是我是 01:40 加入进去队列
//func (e Element[T]) Delay() time.Duration {
//	return e.Timeout
//}

//type DelayQueueV1[T any] struct {
//	q *PriorityQueue[Element[T]]
//}
//
//func UseDelayQueueV1() {
//	q1 := NewDelayQueueV1[Element[int]](10)
//}

//func NewDelayQueueV1[T Element[any]](c int) *DelayQueueV1[T] {
//	return &DelayQueueV1[T]{
//		q: NewPriorityQueue[Element[T]](c, func(src Element[T], dst Element[T]) int {
//			if src.Timeout > dst.Timeout {
//				return 1
//			} else if src.Timeout < dst.Timeout {
//				return -1
//			}
//			return 0
//		}),
//	}
//}

//func NewDelayQueue[T Element[any]](size int) *DelayQueue[T] {
//}
