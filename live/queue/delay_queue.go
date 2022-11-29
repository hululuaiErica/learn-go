package queue

import (
	"context"
	"sync"
	"time"
)

type DelayQueue[T Delayable] struct {
	q *PriorityQueue[T]
	mutex sync.RWMutex
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	return &DelayQueue[T]{
		q: NewPriorityQueue[T](capacity, func(src, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay < dstDelay {
				return -1
			} else if srcDelay == dstDelay {
				return 0
			} else {
				return 1
			}
		}),
	}
}

// 入队和并发阻塞队列没太大区别
func (c *DelayQueue[T]) EnQueue(ctx context.Context, data T) error {
	panic("implement me")
}

// 出队就有讲究了：
// 1. Delay() 返回 <= 0 的时候才能出队
// 2. 如果队首的 Delay()=300ms >0，要是 sleep，等待 Delay() 降下去
// 3. 如果正在 sleep 的过程，有新元素来了，
//    并且 Dealay() = 200 比你正在sleep 的时间还要短，你要调整你的 sleep 时间
// 4. 如果 sleep 的时间还没到，就超时了，那么就返回
// sleep 本质上是阻塞（你可以用 time.Sleep，你也可以用 channel）
func (c *DelayQueue[T]) DeQueue(ctx context.Context) (T, error) {
	panic("implement")
}

type Delayable interface {
	Delay() time.Duration
	// Deadline() time.Time
}
