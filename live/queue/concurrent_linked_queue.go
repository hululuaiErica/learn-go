package queue

import "context"

// ConcurrentLinkedQueue 基于原子操作的并发队列
// 不考虑并发的情况下，怎么实现这个 ConcurrentLinkedQueue
// 改造成原子操作
type ConcurrentLinkedQueue[T any] struct {
	head *node[T]
	tail *node[T]
	zero T
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkedQueue[T] {
	newNode := &node[T]{}
	return &ConcurrentLinkedQueue[T]{
		head: newNode,
		tail: newNode,
	}
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, val T) error {
	newNode := &node[T]{
		data: val,
	}
	// 是这样的吧?
	// 这是一个原子 CAS 操作
	c.tail.next = newNode
	// 这边又是一个原子 CAS 操作
	c.tail = newNode
	return nil
}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// 读取 head 和 tail，你要原子操作
	// 读完之后比较
	if c.head == c.tail {
		return c.zero, ErrEmptyQueue
	}
	head := c.head
	// 这边又是一个 CAS 操作
	headNext := head.next
	return headNext.data, nil
}

type node[T any] struct {
	data T
	next *node[T]
}
