package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testDelayQueue = DelayQueueV3[int, *ElementV3[int]]

func newTestDelayQueueV3(cap int) *testDelayQueue {
	return NewDelayQueueV3[int, *ElementV3[int]](cap)
}

func mustEnqueueV3(val int, delay int64) func(t *testing.T, queue *testDelayQueue) {
	return func(t *testing.T, queue *testDelayQueue) {
		require.NoError(t, queue.Enqueue(context.Background(),
			newTestElmV3(val, delay)))
	}
}

func newTestElmV3(value int, delay int64) *ElementV3[int] {
	return &ElementV3[int]{
		value:    value,
		deadline: time.Now().Add(time.Millisecond * time.Duration(delay)),
	}
}

func TestDelayQueueV3_Enqueue(t *testing.T) {
	type testCase[R int, T DelayableV3[R]] struct {
		name    string
		queue   *DelayQueueV3[R, T]
		before  func(t *testing.T, queue *DelayQueueV3[R, T])
		while   func(t *testing.T, queue *DelayQueueV3[R, T])
		after   func(t *testing.T, queue *DelayQueueV3[R, T])
		value   int
		delay   int64
		timeout int64
		wantErr error
	}
	tests := []testCase[int, *ElementV3[int]]{
		{
			name:  "enqueue to empty queue",
			queue: newTestDelayQueueV3(1),
			after: func(t *testing.T, queue *testDelayQueue) {
				val, err := queue.priorityQueue.Dequeue()
				require.NoError(t, err)
				require.Equal(t, 1, val.value)
			},
			timeout: 10,
			value:   1,
		},
		{
			name:  "enqueue active element to full queue",
			queue: newTestDelayQueueV3(1),
			before: func(t *testing.T, queue *testDelayQueue) {
				mustEnqueueV3(1, 60)(t, queue)
			},
			timeout: 40,
			delay:   20,
			wantErr: ErrOutOfCapacity,
		},
		{
			name:    "enqueue inactive element to full queue",
			queue:   newTestDelayQueueV3(1),
			before:  mustEnqueueV3(1, 60),
			timeout: 20,
			delay:   40,
			wantErr: ErrOutOfCapacity,
		},
		{
			name:   "enqueue to full queue while dequeue valid element",
			queue:  newTestDelayQueueV3(1),
			before: mustEnqueueV3(1, 60),
			while: func(t *testing.T, queue *testDelayQueue) {
				_, err := queue.Dequeue(context.Background())
				require.NoError(t, err)
			},
			timeout: 80,
		},
		{
			name:   "enqueue active element to full queue while dequeue invalid element",
			queue:  newTestDelayQueueV3(1),
			before: mustEnqueueV3(1, 60),
			while: func(t *testing.T, queue *testDelayQueue) {
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
			queue:  newTestDelayQueueV3(1),
			before: mustEnqueueV3(1, 60),
			while: func(t *testing.T, queue *testDelayQueue) {
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
			err := tt.queue.Enqueue(ctx, newTestElmV3(tt.value, tt.delay))
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDelayQueueV3_Dequeue(t *testing.T) {
	type testCase[R int, T DelayableV3[R]] struct {
		name    string
		queue   *DelayQueueV3[R, T]
		before  func(t *testing.T, queue *DelayQueueV3[R, T])
		while   func(t *testing.T, queue *DelayQueueV3[R, T])
		timeout int64
		wantVal int
		wantErr error
	}
	tests := []testCase[int, *ElementV3[int]]{
		{
			name:    "dequeue from empty queue",
			queue:   newTestDelayQueueV3(1),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue new active element from empty queue",
			queue:   newTestDelayQueueV3(1),
			while:   mustEnqueueV3(1, 20),
			timeout: 4000,
			wantVal: 1,
		},
		{
			name:    "dequeue new inactive element from empty queue",
			queue:   newTestDelayQueueV3(1),
			while:   mustEnqueueV3(1, 60),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue active element from full queue",
			queue:   newTestDelayQueueV3(1),
			before:  mustEnqueueV3(1, 60),
			timeout: 80,
			wantVal: 1,
		},
		{
			name:    "dequeue inactive element from full queue",
			queue:   newTestDelayQueueV3(1),
			before:  mustEnqueueV3(1, 60),
			timeout: 20,
			wantErr: ErrEmptyQueue,
		},
		{
			name:    "dequeue new active element from full queue",
			queue:   newTestDelayQueueV3(1),
			before:  mustEnqueueV3(1, 60),
			while:   mustEnqueueV3(2, 40),
			timeout: 80,
			wantVal: 2,
		},
		{
			name:    "dequeue new inactive element from full queue",
			queue:   newTestDelayQueueV3(1),
			before:  mustEnqueueV3(1, 60),
			while:   mustEnqueueV3(2, 40),
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

func BenchmarkDelayQueueV3(b *testing.B) {
	const delay = 0
	const capacity = 100

	b.Run("enqueue", func(b *testing.B) {
		queue := newTestDelayQueueV3(b.N)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = queue.Enqueue(context.Background(), newTestElmV3(1, delay))
		}
	})

	b.Run("parallel to enqueue", func(b *testing.B) {
		queue := newTestDelayQueueV3(b.N)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = queue.Enqueue(context.Background(), newTestElmV3(1, delay))
			}
		})
	})

	b.Run("dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV3(b.N)
		for i := 0; i < b.N; i++ {
			require.NoError(b, queue.Enqueue(context.Background(), newTestElmV3(1, delay)))
		}
		time.Sleep(time.Millisecond * delay)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = queue.Dequeue(context.Background())
		}
	})

	b.Run("parallel to dequeue", func(b *testing.B) {
		queue := newTestDelayQueueV3(b.N)
		for i := 0; i < b.N; i++ {
			require.NoError(b, queue.Enqueue(context.Background(), newTestElmV3(1, delay)))
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
		queue := newTestDelayQueueV3(capacity)
		go func() {
			for i := 0; i < b.N; i++ {
				_ = queue.Enqueue(context.Background(), newTestElmV3(i, delay))
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
		queue := newTestDelayQueueV3(capacity)
		go func() {
			for i := 0; i < b.N; i++ {
				_, _ = queue.Dequeue(context.Background())
			}
		}()
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = queue.Enqueue(context.Background(), newTestElmV3(1, delay))
			}
		})
	})

	b.Run("parallel to enqueue and dequeue", func(b *testing.B) {
		var wg sync.WaitGroup
		var (
			enqueueSeq atomic.Int32
			dequeueSeq atomic.Int32
			checksum   atomic.Int64
		)
		queue := newTestDelayQueueV3(capacity)
		b.ReportAllocs()
		b.ResetTimer()
		procs := runtime.GOMAXPROCS(0)
		wg.Add(procs)
		for i := 0; i < procs; i++ {
			go func(i int) {
				defer wg.Done()
				for {
					if i%2 == 0 {
						if seq := int(enqueueSeq.Add(1)); seq <= b.N {
							for {
								if err := queue.Enqueue(context.Background(), newTestElmV3(seq, delay)); err == nil {
									break
								}
							}
						} else {
							return
						}
					} else {
						if seq := int(dequeueSeq.Add(1)); seq > b.N {
							return
						}
						for {
							if elm, err := queue.Dequeue(context.Background()); err == nil {
								checksum.Add(int64(elm.Value()))
								break
							}
						}
					}
				}
			}(i)
		}
		wg.Wait()
		assert.Zero(b, queue.priorityQueue.Len())
		assert.Equal(b, int64((1+b.N)*b.N/2), checksum.Load())
	})
}
