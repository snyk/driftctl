package cache

import "sync"

type Cache struct {
	mu    *sync.RWMutex
	store map[string]interface{}
}

func New() *Cache {
	return &Cache{
		mu:    &sync.RWMutex{},
		store: map[string]interface{}{},
	}
}

func (c *Cache) Set(key string, value interface{}) (overridden bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[key]; ok {
		overridden = true
	}
	c.store[key] = value
	return
}

func (c *Cache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.store[key]
}
