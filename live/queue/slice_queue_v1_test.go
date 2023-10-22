package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewSliceQueueV1(t *testing.T) {
	q := NewSliceQueueV1[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := q.Enqueue(ctx, 1)
	assert.NoError(t, err)
	go func() {
		for {
			ctx1, cancel1 := context.WithTimeout(context.Background(),
				time.Second*3)
			defer cancel1()
			val, err1 := q.Dequeue(ctx1)
			println(val, err1)
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
