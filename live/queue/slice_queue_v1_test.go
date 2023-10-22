package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNewSliceQueueV1(t *testing.T) {
	q := NewSliceQueueV1[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := q.Enqueue(ctx, 1)
	// 这一步，writeCh 是不是有一个元素？也就是 writeChan 满了
	assert.NoError(t, err)
	go func() {
		for {
			ctx1, cancel1 := context.WithTimeout(context.Background(),
				time.Second*3)
			defer cancel1()
			val, err1 := q.Dequeue(ctx1)
			t.Log(val, err1)
		}
	}()
	go func() {
		http.ListenAndServe(":8081", nil)
	}()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = q.Enqueue(ctx, 2)
	assert.NoError(t, err)
}

func TestNewSliceQueueV1_Dequeue2(t *testing.T) {
	q := NewSliceQueueV1[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := q.Enqueue(ctx, 1)
	// 这一步，writeCh 是不是有一个元素？也就是 writeChan 满了
	assert.NoError(t, err)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err := q.Dequeue(ctx)
	t.Log(val, err)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err = q.Dequeue(ctx)
	t.Log(val, err)
}

func TestBroadcast(t *testing.T) {
	l := &sync.Mutex{}
	cond := sync.NewCond(l)
	cond.Broadcast()
	// 必然崩溃
	t.Log("abc")
}
