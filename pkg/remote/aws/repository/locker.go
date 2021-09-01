package repository

import "sync"

type methodLocker struct {
	h map[string]sync.Locker
	l sync.Locker
}

func newMethodLocker() *methodLocker {
	return &methodLocker{
		h: map[string]sync.Locker{},
		l: &sync.Mutex{},
	}
}

func (l *methodLocker) Lock(method string) {
	l.findOrCreate(method).Lock()
}

func (l *methodLocker) Unlock(method string) {
	l.findOrCreate(method).Unlock()
}

func (l *methodLocker) findOrCreate(method string) sync.Locker {
	if _, ok := l.h[method]; !ok {
		l.l.Lock()
		l.h[method] = &sync.Mutex{}
		l.l.Unlock()
	}
	return l.h[method]
}
