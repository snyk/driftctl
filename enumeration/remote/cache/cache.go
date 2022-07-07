package cache

import (
	"container/list"
	"sync"
)

type Cache interface {
	Put(string, interface{}) bool
	Get(string) interface{}
	GetAndLock(string) interface{}
	Unlock(string)
	Len() int
}

type LRUCache struct {
	cap     int
	mu      *sync.Mutex
	l       *list.List
	m       map[string]*list.Element
	lockMap *sync.Map
}

type pair struct {
	key   string
	value interface{}
}

func New(capacity int) Cache {
	return &LRUCache{
		cap:     capacity,
		mu:      &sync.Mutex{},
		l:       &list.List{},
		m:       make(map[string]*list.Element, capacity),
		lockMap: &sync.Map{},
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

func (c *LRUCache) Put(key string, value interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cap == 0 {
		return false
	}

	// if the key already exists, move to front and update the value
	if node, ok := c.m[key]; ok {
		c.l.MoveToFront(node)
		node.Value.(*list.Element).Value = pair{key: key, value: value}
		return true
	}

	// if the list is full, delete the last element
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

	return false
}

func (c *LRUCache) Len() int {
	return c.l.Len()
}

func (c *LRUCache) GetAndLock(s string) interface{} {
	lock, _ := c.lockMap.LoadOrStore(s, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	return c.Get(s)
}

func (c *LRUCache) Unlock(s string) {
	lock, exist := c.lockMap.Load(s)
	if exist {
		lock.(*sync.Mutex).Unlock()
	}
}
