package queue

import (
	"context"
	"sync/atomic"
	"unsafe"
)

// ConcurrentLinkedQueue 基于原子操作的并发队列
// 不考虑并发的情况下，怎么实现这个 ConcurrentLinkedQueue
// 改造成原子操作
type ConcurrentLinkedQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	zero T
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkedQueue[T] {
	newNode := &node[T]{}
	ptr := unsafe.Pointer(newNode)
	return &ConcurrentLinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, val T) error {
	newNode := &node[T]{
		data: val,
	}
	newNodePtr := unsafe.Pointer(newNode)
	// tail := c.tail
	for {
		tailPtr := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPtr)
		//tailNextPtr := atomic.LoadPointer(&tail.next)
		//if tailNextPtr != nil {
		//	continue
		//}

		// 这是一个原子 CAS 操作
		if atomic.CompareAndSwapPointer(&tail.next, nil, newNodePtr) {
			// 说明你入队成功了

			// 可能失败，失败意味着，后面的 goroutine 已经修正了 c.tail
			atomic.CompareAndSwapPointer(&c.tail, tailPtr, newNodePtr)
			return nil
		}
	}
	//tail.next = newNode
	// 这边又是一个原子 CAS 操作
	//c.tail = newNode
}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// 读取 head 和 tail，你要原子操作
	// 读完之后比较
	// 同一时刻的有 N 个 goroutine，最终 for 循环的次数应该是
	// n + (n-1) + (n-2) + (n-3) + ... + 1
	// n(n+1)/2
	for {
		// LongAdder
		headPtr := atomic.LoadPointer(&c.head)
		if headPtr == atomic.LoadPointer(&c.tail) {
			return c.zero, ErrEmptyQueue
		}
		head := (*node[T])(headPtr)
		headNextPtr := atomic.LoadPointer(&head.next)
		// 这边又是一个 CAS 操作
		if atomic.CompareAndSwapPointer(&c.head, headPtr, headNextPtr) {
			headNext := (*node[T])(headNextPtr)
			return headNext.data, nil
		}
	}
}

type node[T any] struct {
	data T
	next unsafe.Pointer
}
