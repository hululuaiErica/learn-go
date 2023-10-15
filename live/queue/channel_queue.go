package queue

type ChannelQueue[T any] struct {
	ch   chan T
	head T
	zero T
}

func (q *ChannelQueue[T]) Peek() T {
	return q.head
}

func (q *ChannelQueue[T]) Pop() T {
	// 用原子操作
	//head := q.head
	//if head == q.zero {
	//	select {
	//	case val := <-q.ch:
	//		return val
	//	default:
	//
	//	}
	//}
	//
	//select {
	//case val := <-q.ch:
	//	q.head = val
	//default:
	//	q.head = q.zero
	//}
	//return head
	panic("implement me")
}

type QueueDemo interface {
	Enqueue(val any)
}

type PersistentQueue struct {
	q QueueDemo
}

func (q *PersistentQueue) Enqueue(val any) {
	// 在这里持久化
	q.q.Enqueue(val)
}
