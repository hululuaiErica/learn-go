package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"math/rand"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	weight := semaphore.NewWeighted(10)
	ch := make(chan any, 1)
	go func() {
		err := weight.Acquire(context.Background(), 1)
		t.Log(err)
		ch <- err
	}()
	<-ch
}

func TestSliceQueue_In(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		in   int
		q    *SliceQueue[int]

		wantErr  error
		wantData []int
	}{
		{
			name: "超时",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			}(),
			in: 10,
			q: func() *SliceQueue[int] {
				q := NewSliceQueue[int](2)
				_ = q.In(context.Background(), 11)
				_ = q.In(context.Background(), 12)
				return q
			}(),
			wantErr:  context.DeadlineExceeded,
			wantData: []int{11, 12},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.q.In(tc.ctx, tc.in)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, tc.q.data)
		})
	}
}

func TestSliceQueue_Out(t *testing.T) {

}

func TestSliceQueue_InOut(t *testing.T) {
	// 这边几十个 goroutine 入队

	// 这边几十个 goroutine 出队

	q := NewSliceQueue[int](10)
	closed := false
	for i := 0; i < 20; i++ {
		go func() {
			for {
				if closed {
					return
				}
				val := rand.Int()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_ = q.In(ctx, val)
				// 如何校验 err 的值？
				cancel()
			}
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			for {
				if closed {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.Out(ctx)
				cancel()
				// 如何 校验 val 的值，和 err 的值？
			}
		}()
	}

	time.Sleep(time.Second * 10)
	closed = true
}
