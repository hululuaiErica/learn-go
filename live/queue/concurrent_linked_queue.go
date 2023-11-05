package queue

import "context"

// ConcurrentLinkedQueue 基于原子操作的并发队列
// 不考虑并发的情况下，怎么实现这个 ConcurrentLinkedQueue
type ConcurrentLinkedQueue[T any] struct {
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, val T) error {

	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	//TODO implement me
	panic("implement me")
}

type node[T any] struct {
	data *T
	next *node[T]
}
