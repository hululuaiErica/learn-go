package queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"testing"
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

}

func TestSliceQueue_Out(t *testing.T) {

}

func TestSliceQueue_InOut(t *testing.T) {
	// 这边几十个 goroutine 入队

	// 这边几十个 goroutine 出队

	// 如何验证结果的正确性？
}