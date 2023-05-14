package queue

import (
	"errors"
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
	//head := &node[T]{}
	//head.next = head
	return &LinkedQueue[T]{
		//head: head,
		//tail: head,
	}
}

func (l *LinkedQueue[T]) In(val T) error {
	newNode := &node[T]{data: val}
	newPtr := unsafe.Pointer(newNode)
	for {
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*node[T])(tailPtr)
		tailNext := atomic.LoadPointer(&tail.next)
		if tailNext != nil {
			continue // 被人并发修改了
		}
		// 先指向新节点
		// 再调整 tail 节点
		if atomic.CompareAndSwapPointer(&tail.next, tailNext, newPtr) {
			atomic.CompareAndSwapPointer(&l.tail, tailPtr, newPtr)
		}
	}
}

func (l *LinkedQueue[T]) Out() (T, error) {
	for {
		headPtr := atomic.LoadPointer(&l.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*node[T])(tailPtr)

		if head == tail {
			return l.zero, errors.New("empty queue")
		}

		headNextPtr := atomic.LoadPointer(&head.next)
		if atomic.CompareAndSwapPointer(&l.head, headPtr, headNextPtr) {
			headNext := (*node[T])(headNextPtr)
			return headNext.data, nil
		}
	}
}

type node[T any] struct {
	data T
	// *node[T]
	next unsafe.Pointer
}
