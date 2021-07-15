package memstore

import (
	"sync"
)

type Bucket interface {
	Set(string, interface{})
	Get(string) interface{}
	Values() map[string]interface{}
}

type bucket struct {
	m      *sync.RWMutex
	values map[string]interface{}
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

func (b bucket) Values() map[string]interface{} {
	return b.values
}
