package interview

import (
	"sync"
)

type KeyLock struct {
	locks sync.Map
}

func (m *KeyLock) Lock(key string) {
	lock, _ := m.locks.LoadOrStore(key, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
}

func (m *KeyLock) UnLock(key string) {
	lock, _ := m.locks.LoadOrStore(key, &sync.Mutex{})
	lock.(*sync.Mutex).Unlock()
}
