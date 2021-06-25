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
	m       *sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	m      *sync.RWMutex
	values map[string]interface{}
}

func New() Store {
	return &store{
		m:       &sync.Mutex{},
		buckets: map[string]*bucket{},
	}
}

func (s store) Bucket(name string) Bucket {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.buckets[name]; !ok {
		s.buckets[name] = &bucket{
			m:      &sync.RWMutex{},
			values: map[string]interface{}{},
		}
	}

	return s.buckets[name]
}

func (b bucket) Set(key string, value interface{}) {
	b.m.Lock()
	defer b.m.Unlock()
	b.values[key] = value
}

func (b bucket) Get(key string) interface{} {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.values[key]
}

func (b bucket) MarshallJSON() ([]byte, error) {
	return json.Marshal(b.values)
}
