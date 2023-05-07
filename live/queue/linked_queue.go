package queue

// 基于链表的，非阻塞的，使用锁的并发队列

type LinkedQueue[T any] struct {
}

func (l *LinkedQueue[T]) In(val T) error {
	//TODO implement me
	panic("implement me")
}

func (l *LinkedQueue[T]) Out() (T, error) {
	//TODO implement me
	panic("implement me")
}
