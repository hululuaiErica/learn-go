package queue

import "context"

// 基于链表的，非阻塞的，使用锁的并发队列

type LinkedQueue[T any] struct {
}

func (l *LinkedQueue[T]) In(ctx context.Context, val T) error {
	//TODO implement me
	panic("implement me")
}

func (l *LinkedQueue[T]) Out(ctx context.Context) (T, error) {
	//TODO implement me
	panic("implement me")
}

