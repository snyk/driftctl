package cache

import (
	"sync"
)

type Cache interface {
	Set(string, interface{}) bool
	Get(string) interface{}
}

type Store struct {
	mu    *sync.RWMutex
	store map[string]interface{}
}

func New() Cache {
	return &Store{
		mu:    &sync.RWMutex{},
		store: map[string]interface{}{},
	}
}

func (c *Store) Set(key string, value interface{}) (overridden bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[key]; ok {
		overridden = true
	}
	c.store[key] = value
	return
}

func (c *Store) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.store[key]
}
