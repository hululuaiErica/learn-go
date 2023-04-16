package class1

import (
	"context"
	"sync"
)

type Cond struct {
	Cond *sync.Cond
}

func (m *Cond) Wait(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	ch := make(chan struct{})
	go func() {
		m.Cond.Wait()
		m.Cond.L.Unlock()
		select {
		case ch <- struct{}{}:
		default:
			m.Cond.L.Lock()
			m.Cond.Signal()
			m.Cond.L.Unlock()
		}
	}()
	m.Cond.L.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

func (m *Cond) Signal() {
	m.Cond.Signal()
}
