package queue

import "context"

type DelayQueue[T any] struct {
	q *PriorityQueue[T]
}

func (d *DelayQueue[T]) In(ctx context.Context, val T) error {
	//TODO implement me
	panic("implement me")
}

// 出队永远拿到"到期"了的
// 如果没有到期的元素，就阻塞，直到有元素到期
// 如果超时了，直接返回

// 先考虑 Out，你会把代码写成什么样子
func (d *DelayQueue[T]) Out(ctx context.Context) (T, error) {
	//TODO implement me
	panic("implement me")
}
