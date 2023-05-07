package queue

import (
	"sync/atomic"
	"unsafe"
)

// LinkedQueue 基于链表的，非阻塞的，不使用锁
// 用 CAS 操作去修改 LinkedQueue
type LinkedQueue[T any] struct {
	// *node[T]
	head unsafe.Pointer
	// *node[T]
	tail unsafe.Pointer
	zero T
}

func NewLinkedQueue[T any]() *LinkedQueue[T] {
	head := &node[T]{}
	head.next = head
	return &LinkedQueue[T]{
		head: head,
		tail: head,
	}
}

func (l *LinkedQueue[T]) In(val T) error {
	newNode := &node[T]{data: val}
	newPtr := unsafe.Pointer(newNode)
	for {
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*node[T])(tailPtr)
		tailNext := atomic.LoadPointer(&tail.next)

	}
}

func (l *LinkedQueue[T]) Out() (T, error) {
	for {
		headPtr := atomic.LoadPointer(&l.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*node[T])(tailPtr)
	}
}

type node[T any] struct {
	data T
	// *node[T]
	next unsafe.Pointer
}
