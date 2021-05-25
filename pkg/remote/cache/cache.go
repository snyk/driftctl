package cache

import (
	"container/list"
	"sync"
)

type Cache interface {
	Put(string, interface{}) bool
	Get(string) interface{}
	Len() int
}

type LRUCache struct {
	cap int
	mu  *sync.Mutex
	l   *list.List
	m   map[string]*list.Element
}

type pair struct {
	key   string
	value interface{}
}

func New(capacity int) Cache {
	return &LRUCache{
		cap: capacity,
		mu:  &sync.Mutex{},
		l:   &list.List{},
		m:   make(map[string]*list.Element, capacity),
	}
}

func (c *LRUCache) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if the key exists, move to front
	if node, ok := c.m[key]; ok {
		val := node.Value.(*list.Element).Value.(pair).value
		c.l.MoveToFront(node)
		return val
	}
	return nil
}

func (c *LRUCache) Put(key string, value interface{}) (overridden bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if the key already exists, move to front and update the value
	if node, ok := c.m[key]; ok {
		c.l.MoveToFront(node)
		node.Value.(*list.Element).Value = pair{key: key, value: value}
		overridden = true
		return
	}

	// delete the last list node if the list is full
	if c.l.Len() == c.cap {
		idx := c.l.Back().Value.(*list.Element).Value.(pair).key
		delete(c.m, idx)
		c.l.Remove(c.l.Back())
	}

	// initialize a new list node
	node := &list.Element{
		Value: pair{
			key:   key,
			value: value,
		},
	}
	element := c.l.PushFront(node)
	c.m[key] = element

	return
}

func (c *LRUCache) Len() int {
	return c.l.Len()
}
