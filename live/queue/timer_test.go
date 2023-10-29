package queue

import (
	"testing"
	"time"
)

func TestTimerTest(t *testing.T) {
	timer := time.NewTimer(time.Minute)
	start := time.Now()
	go func() {
		<-timer.C
		t.Log("hello, world", start.Second(), time.Now().Second())
	}()
	time.Sleep(time.Second)
	timer.Reset(0)
	time.Sleep(time.Second * 3)
}
