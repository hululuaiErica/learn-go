package queue

import (
	"context"
)

type Queue[T any] interface {
	// 我的方法该怎么定义？
	// 有什么方法，每个方法的输入输出应该是什么？

	Enqueue(ctx context.Context, val T) error
	Dequeue(ctx context.Context) (T, error)

	// 要不要有返回值？返回值究竟是什么？
	// 一边遍历，一边修改队列本身（增加或者删除元素），怎么搞？
	//Iterate(ctx context.Context, f func(val T) bool)
	//Iterate(ctx context.Context, f func(idx int, val T) bool)

	//IterateV1(ctx context.Context,
	//	f func(val T) error) error

	//Iterator() (Cursor[T], error)
}

type MyQueue[T any] struct {
	data []T
}

func (q *MyQueue[T]) Iterate(ctx context.Context,
	f func(val T) bool) {
	for _, val := range q.data {
		ctn := f(val)
		// 返回 bool 值如果是 false，则直接中断遍历
		if !ctn {
			return
		}
	}
}

func (q *MyQueue[T]) IterateV1(ctx context.Context,
	f func(val T) error) error {
	for _, val := range q.data {
		err := f(val)
		// 返回 bool 值如果是 false，则直接中断遍历
		if err != nil {
			return err
		}
	}
	return nil
}

type Cursor[T any] interface {
	Next() bool
	Cur() T
	Err() error
	// 修改到 Queue 本身
	Delete() error
	Append() error
}

type User struct {
	Name string
}

//func UseQueue(q Queue[any]) {
//	cursor, _ := q.Iterator()
//	for cursor.Next() {
//		val := cursor.Cur()
//		// 你就可以随便用
//		fmt.Println(val)
//	}
//	// 这边是要要么遍历完成了，要么是出错了
//	if cursor.Err() != nil {
//		// 说明遍历的过程出错了
//	}
//}
