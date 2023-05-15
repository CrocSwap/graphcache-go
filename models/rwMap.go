package models

import "sync"

type RWLockMap[Key comparable, Val any] struct {
	entries map[Key]Val
	lock    sync.RWMutex
}

func (m *RWLockMap[Key, Val]) lookup(key Key) (Val, bool) {
	m.lock.RLock()
	result, ok := m.entries[key]
	m.lock.Unlock()
	return result, ok
}

func newRwLockMap[Key comparable, Val any]() RWLockMap[Key, Val] {
	return RWLockMap[Key, Val]{
		entries: make(map[Key]Val),
		lock:    sync.RWMutex{},
	}
}
