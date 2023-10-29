package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

type DelayQueueV1[T DelayableV1] struct {
	q             *PriorityQueue[T]
	lock          sync.RWMutex
	dequeueSignal *cond
	enqueueSignal *cond
	zero          T

	expiredCh chan T
	timer     *time.Timer
}

// 写到 21:00
func (d *DelayQueueV1[T]) Enqueue(ctx context.Context, val T) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		d.lock.Lock()
		err := d.q.Enqueue(val)
		switch err {
		case ErrOutOfCapacity:
			// signalCh 已经释放了锁
			signal := d.dequeueSignal.signalCh()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-signal:
			}
		case nil:
			// broadcast 已经释放了锁
			d.enqueueSignal.broadcast()
			return nil
		default:
			d.lock.Unlock()
			// 未知错误
			return err
		}
	}
}

func (d *DelayQueueV1[T]) DequeueV3(ctx context.Context) (T, error) {
	ch := d.DequeueV2()
	select {
	case t, ok := <-ch:
		if !ok {
			return d.zero, errors.New("队列被关闭了")
		}
		return t, nil
	case <-ctx.Done():
		return d.zero, ctx.Err()
	}
}

func (d *DelayQueueV1[T]) EnqueueV2(ctx context.Context, val T) error {
	for {
		d.lock.Lock()
		err := d.q.Enqueue(val)
		switch err {
		case nil:
			// 奇技淫巧
			d.timer.Reset(0)
			d.lock.Unlock()
			return nil
		case ErrOutOfCapacity:
			d.lock.Unlock()
			select {
			case <-ctx.Done():
			case <-d.timer.C:

			}
		}
	}
}

func (d *DelayQueueV1[T]) DequeueV2() <-chan T {
	return d.expiredCh
}

func (d *DelayQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	//TODO implement me
	panic("implement me")
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

func NewDelayQueueV1_Ch[T DelayableV1](c int) *DelayQueueV1[T] {
	timer := time.NewTimer(time.Second)
	timer.Stop()
	res := &DelayQueueV1[T]{
		expiredCh: make(chan T),
		timer:     timer,
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

	go func() {
		for {
			res.lock.Lock()
			// 在这里就是不断检测队首元素，然后还是要监听入队元素
			val, err := res.q.Peek()
			switch err {
			case nil:
				dur := val.Deadline().Sub(time.Now())
				if dur > 0 {
					timer.Reset(dur)
					res.lock.Unlock()
					select {
					case <-timer.C:
					}
				} else {
					val, _ = res.q.Dequeue()
					res.expiredCh <- val
					// 还要有一个信号
					res.timer.Reset(0)
				}
			case ErrEmptyQueue:
			default:
				panic("不可能的分支")
			}
		}
	}()

	return res
}

func NewDelayQueueV1[T DelayableV1](c int) *DelayQueueV1[T] {
	return &DelayQueueV1[T]{
		// 这里要不要 buffer，要多大？
		//enqueueCh: make(chan struct{}, c),
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

type cond struct {
	signal chan struct{}
	l      sync.Locker
}

func newCond(l sync.Locker) *cond {
	return &cond{
		signal: make(chan struct{}),
		l:      l,
	}
}

// broadcast 唤醒等待者
// 如果没有人等待，那么什么也不会发生
// 必须加锁之后才能调用这个方法
// 广播之后锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) broadcast() {
	signal := make(chan struct{})
	old := c.signal
	c.signal = signal
	c.l.Unlock()
	close(old)
}

// signalCh 返回一个 channel，用于监听广播信号
// 必须在锁范围内使用
// 调用后，锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) signalCh() <-chan struct{} {
	res := c.signal
	c.l.Unlock()
	return res
}

// 到 20:20 分，大家写一下
//func (d *DelayQueueV1[T]) DequeueV1(ctx context.Context) (T, error) {
//	err := d.dequeueCap.Acquire(ctx, 1)
//	if err != nil {
//		return d.zero, err
//	}
//	d.lock.Lock()
//
//	if ctx.Err() != nil {
//		d.dequeueCap.Release(1)
//		d.lock.Unlock()
//		return d.zero, ctx.Err()
//	}
//
//	// dequeueCap 可以确保，你这里肯定有
//	val, _ := d.q.Peek()
//	ddl := val.Deadline()
//	now := time.Now()
//	if ddl.Before(now) {
//		d.enqueueCap.Release(1)
//		d.lock.Unlock()
//		return val, nil
//	}
//
//	timer := time.NewTimer(ddl.Sub(now))
//
//	// 要先释放锁
//	d.lock.Unlock()
//	// 这边就是要考虑等val的过期时间
//	// 还有 ctx 超时时间
//	select {
//	case <-ctx.Done():
//		d.dequeueCap.Release(1)
//		return d.zero, ctx.Err()
//	case <-timer.C:
//		return val, nil
//		//case <-d.enqueueCh: // 有新元素来了的分支
//
//	}
//}

// Enqueue 先来实现
//func (d *DelayQueueV1[T]) EnqueueV1(ctx context.Context, val T) error {
//	//err := d.enqueueCap.Acquire(ctx, 1)
//	if err != nil {
//		// 这边就是满了
//		return err
//	}
//	d.lock.Lock()
//	defer d.lock.Unlock()
//	if ctx.Err() != nil {
//		//d.enqueueCap.Release(1)
//		return ctx.Err()
//	}
//
//	err = d.q.Enqueue(val)
//	// 基本不可能的分支
//	if err != nil {
//		return err
//	}
//	// 发个信号
//	//select {
//	//case d.enqueueCh <- struct{}{}:
//	//default:
//	//
//	//}
//	//d.dequeueCap.Release(1)
//	return nil
//}

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
