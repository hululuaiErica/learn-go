package queue

import (
	"context"
)

type Queue[T any] interface {
	// Push(msg T) error
	// Pop() (msg T, err error)
	// IsEmpty() bool
	// Clear() error
	// Close() error
	// Size() int

	// Capacity() int

	// Get(ctx context.Context, index int) (T, error)
	// Set(ctx context.Context, index int, data T) error
	// Add(ctx context.Context, index int, data T) error


	// Queue[User] => 放的是结构体
	// Queue[*User] => 放的就是指针
	// In(context.Background()) 这样永不超时
	In(ctx context.Context, val T) error
	Out(ctx context.Context) (T, error)

	// 瞬时的
	// IsEmpty() bool


	// timeoutUnit 可以是毫秒，秒，纳秒，分钟，小时
	// InV1(timeout int64, timeoutUnit int8, val T) error

	// InV2(timeout time.Duration, val T) error
	// Out(ctx context.Context) (T, error)
	// IsEmpty() bool
}