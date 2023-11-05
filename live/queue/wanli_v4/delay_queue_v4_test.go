package queue

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type testDelayQueueV4 = DelayQueueV4[int, *ElementV4[int]]

func newTestDelayQueueV4(cap int) *testDelayQueueV4 {
	return NewDelayQueueV4[int, *ElementV4[int]](cap)
}

func mustEnqueueV4(val int, delay int64) func(t *testing.T, queue *testDelayQueueV4) {
	return func(t *testing.T, queue *testDelayQueueV4) {
		require.NoError(t, queue.Enqueue(context.Background(),
			newTestElmV4(val, delay)))
	}
}

func newTestElmV4(value int, delay int64) *ElementV4[int] {
	return &ElementV4[int]{
		value:    value,
		deadline: time.Now().Add(time.Millisecond * time.Duration(delay)),
	}
}

func TestDelayQueueV4_Enqueue(t *testing.T) {
	type testCase[R int, T DelayableV4[R]] struct {
		name    string
		queue   *DelayQueueV4[R, T]
		before  func(t *testing.T, queue *DelayQueueV4[R, T])
		while   func(t *testing.T, queue *DelayQueueV4[R, T])
		after   func(t *testing.T, queue *DelayQueueV4[R, T])
		value   int
		delay   int64
		timeout int64
		wantErr error
	}
	tests := []testCase[int, *ElementV4[int]]{
		{
			name:  "enqueue to empty queue",
			queue: newTestDelayQueueV4(1),
			after: func(t *testing.T, queue *testDelayQueueV4) {
				val, err := queue.priorityQueue.Dequeue()
				require.NoError(t, err)
				require.Equal(t, 1, val.value)
			},
			timeout: 10,
			value:   1,
		},
		{
			name:    "enqueue active element to full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			timeout: 40,
			delay:   20,
			wantErr: ErrOutOfCapacity,
		},
		{
			name:    "enqueue inactive element to full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			timeout: 20,
			delay:   40,
			wantErr: ErrOutOfCapacity,
		},
		{
			name:   "enqueue to full queue while dequeue valid element",
			queue:  newTestDelayQueueV4(1),
			before: mustEnqueueV4(1, 60),
			while: func(t *testing.T, queue *testDelayQueueV4) {
				_, err := queue.Dequeue(context.Background())
				require.NoError(t, err)
			},
			timeout: 80,
		},
		{
			name:   "enqueue active element to full queue while dequeue invalid element",
			queue:  newTestDelayQueueV4(1),
			before: mustEnqueueV4(1, 60),
			while: func(t *testing.T, queue *testDelayQueueV4) {
				elm, err := queue.Dequeue(context.Background())
				require.NoError(t, err)
				require.Equal(t, 2, elm.value)
			},
			timeout: 40,
			value:   2,
			delay:   20,
		},
		{
			name:   "enqueue inactive element to full queue while dequeue invalid element",
			queue:  newTestDelayQueueV4(1),
			before: mustEnqueueV4(1, 60),
			while: func(t *testing.T, queue *testDelayQueueV4) {
				_, err := queue.Dequeue(context.Background())
				require.NoError(t, err)
			},
			timeout: 20,
			delay:   40,
			wantErr: ErrOutOfCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Millisecond*time.Duration(tt.timeout))
			defer cancel()
			if tt.before != nil {
				tt.before(t, tt.queue)
			}
			if tt.while != nil {
				go tt.while(t, tt.queue)
			}
			err := tt.queue.Enqueue(ctx, newTestElmV4(tt.value, tt.delay))
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDelayQueueV4_Dequeue(t *testing.T) {
	type testCase[R int, T DelayableV4[R]] struct {
		name    string
		queue   *DelayQueueV4[R, T]
		before  func(t *testing.T, queue *DelayQueueV4[R, T])
		while   func(t *testing.T, queue *DelayQueueV4[R, T])
		timeout int64
		wantVal int
		wantErr error
	}
	tests := []testCase[int, *ElementV4[int]]{
		{
			name:    "dequeue from empty queue",
			queue:   newTestDelayQueueV4(1),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue new active element from empty queue",
			queue:   newTestDelayQueueV4(1),
			while:   mustEnqueueV4(1, 20),
			timeout: 40,
			wantVal: 1,
		},
		{
			name:    "dequeue new inactive element from empty queue",
			queue:   newTestDelayQueueV4(1),
			while:   mustEnqueueV4(1, 60),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue active element from full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			timeout: 80,
			wantVal: 1,
		},
		{
			name:    "dequeue inactive element from full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue new active element from full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			while:   mustEnqueueV4(2, 40),
			timeout: 80,
			wantVal: 2,
		},
		{
			name:    "dequeue new inactive element from full queue",
			queue:   newTestDelayQueueV4(1),
			before:  mustEnqueueV4(1, 60),
			while:   mustEnqueueV4(2, 40),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Millisecond*time.Duration(tt.timeout))
			defer cancel()
			if tt.before != nil {
				tt.before(t, tt.queue)
			}
			if tt.while != nil {
				go tt.while(t, tt.queue)
			}
			got, err := tt.queue.Dequeue(ctx)
			require.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			require.Equal(t, tt.wantVal, got.value)
		})
	}
}

func BenchmarkDelayQueueV4(b *testing.B) {
	const delay = 0
	const capacity = 100

	b.Run("enqueue", func(b *testing.B) {
		queue := newTestDelayQueueV4(b.N)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = queue.Enqueue(context.Background(), newTestElmV4(1, delay))
		}
	})

	b.Run("parallel to enqueue", func(b *testing.B) {
		queue := newTestDelayQueueV4(b.N)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = queue.Enqueue(context.Background(), newTestElmV4(1, delay))
			}
		})
	})

	b.Run("dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV4(b.N)
		for i := 0; i < b.N; i++ {
			require.NoError(b, queue.Enqueue(context.Background(), newTestElmV4(1, delay)))
		}
		time.Sleep(time.Millisecond * delay)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = queue.Dequeue(context.Background())
		}
	})

	b.Run("parallel to dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV4(b.N)
		for i := 0; i < b.N; i++ {
			require.NoError(b, queue.Enqueue(context.Background(), newTestElmV4(1, delay)))
		}
		time.Sleep(time.Millisecond * delay)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = queue.Dequeue(context.Background())
			}
		})
	})

	b.Run("parallel to dequeue while enqueue", func(b *testing.B) {
		queue := newTestDelayQueueV4(capacity)
		go func() {
			for i := 0; i < b.N; i++ {
				_ = queue.Enqueue(context.Background(), newTestElmV4(1, delay))
			}
		}()
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = queue.Dequeue(context.Background())
			}
		})
	})

	b.Run("parallel to enqueue while dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV4(capacity)
		go func() {
			for i := 0; i < b.N; i++ {
				_, _ = queue.Dequeue(context.Background())
			}
		}()
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = queue.Enqueue(context.Background(), newTestElmV4(1, delay))
			}
		})
	})

	b.Run("parallel to enqueue and dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV4(capacity)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = queue.Enqueue(context.Background(), newTestElmV4(1, delay))
				_, _ = queue.Dequeue(context.Background())
			}
		})
	})
}
