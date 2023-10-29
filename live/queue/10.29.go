package queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"time"
)

type DelayQueueTimer[T DelayableV1] struct {
	timer time.Timer

	// 结合这两个来
	enqueueCap *semaphore.Weighted
	dequeueCap *semaphore.Weighted
}

func (d *DelayQueueTimer[T]) Enqueue(ctx context.Context, val T) error {
	// 这边 Enqueue 的时候要借助 timer 来通知 Dequeue
	//TODO implement me
	panic("implement me")
}

func (d *DelayQueueTimer[T]) Dequeue(ctx context.Context) (T, error) {
	// Dequeue 这边要借助 time.Timer 来监听入队，或者队首元素超时
	//TODO implement me
	panic("implement me")
}
