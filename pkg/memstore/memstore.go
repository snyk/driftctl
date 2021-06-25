package memstore

import (
	"encoding/json"
	"sync"
)

type Bucket interface {
	Set(string, interface{})
	Get(string) interface{}
	MarshallJSON() ([]byte, error)
}

type Store interface {
	Bucket(string) Bucket
}

type store struct {
	m *sync.Mutex
	s map[string]*bucket
}

type bucket struct {
	m *sync.RWMutex
	s map[string]interface{}
}

func New() Store {
	return &store{
		m: &sync.Mutex{},
		s: map[string]*bucket{},
	}
}

func (s store) Bucket(name string) Bucket {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.s[name]; !ok {
		s.s[name] = &bucket{
			m: &sync.RWMutex{},
			s: map[string]interface{}{},
		}
	}

	return s.s[name]
}

func (b bucket) Set(key string, value interface{}) {
	b.m.Lock()
	defer b.m.Unlock()
	b.s[key] = value
}

func (b bucket) Get(key string) interface{} {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.s[key]
}

func (b bucket) MarshallJSON() ([]byte, error) {
	return json.Marshal(b.s)
}
