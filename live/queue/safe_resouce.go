package queue

import (
	"log"
	"sync"
)

type SafeResource1 struct {
	mutex sync.Mutex
	data  []any
}

func (s *SafeResource1) AddV1(val any) {
	s.mutex.Lock()
	//defer s.mutex.Unlock()
	s.data = append(s.data, val)
}

func (s SafeResource1) AddV2(val any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	log.Printf("%p", &s.mutex)
	s.data = append(s.data, val)
}

type SafeResource2 struct {
	mutex *sync.Mutex
	data  []any
}

func (s *SafeResource2) AddV1(val any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = append(s.data, val)
}

func (s SafeResource2) AddV2(val any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = append(s.data, val)
}
