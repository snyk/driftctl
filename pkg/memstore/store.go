package memstore

import (
	"sync"
)

type Store interface {
	Bucket(BucketName) Bucket
}

type store struct {
	m       *sync.Mutex
	buckets map[int]*bucket
}

func New() Store {
	return &store{
		m:       &sync.Mutex{},
		buckets: map[int]*bucket{},
	}
}

func (s store) Bucket(name BucketName) Bucket {
	s.m.Lock()
	defer s.m.Unlock()

	key := int(name)
	if _, exist := s.buckets[key]; !exist {
		s.buckets[key] = &bucket{
			m:      &sync.RWMutex{},
			values: map[string]interface{}{},
		}
	}

	return s.buckets[key]
}
