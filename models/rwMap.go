package models

import "sync"

type RWLockMap[Key comparable, Val any] struct {
	entries map[Key]Val
	lock    sync.RWMutex
}

type RWLockMapArray[Key comparable, Val any] struct {
	entries map[Key][]Val
	lock    sync.RWMutex
}

func (m *RWLockMap[Key, Val]) lookup(key Key) (Val, bool) {
	m.lock.RLock()
	result, ok := m.entries[key]
	m.lock.RUnlock()
	return result, ok
}

func (m *RWLockMapArray[Key, Val]) lookup(key Key) ([]Val, bool) {
	m.lock.RLock()
	result, ok := m.entries[key]
	m.lock.RUnlock()
	return result, ok
}

func (m *RWLockMap[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	m.entries[key] = val
	m.lock.Unlock()
}

func (m *RWLockMapArray[Key, Val]) insertList(key Key, val Val) {
	m.lock.Lock()
	result, ok := m.entries[key]
	if !ok {
		m.entries[key] = make([]Val, 0)
	}
	m.entries[key] = append(result, val)
	m.lock.Unlock()
}

func newRwLockMap[Key comparable, Val any]() RWLockMap[Key, Val] {
	return RWLockMap[Key, Val]{
		entries: make(map[Key]Val),
		lock:    sync.RWMutex{},
	}
}

func newRwLockMapArray[Key comparable, Val any]() RWLockMapArray[Key, Val] {
	return RWLockMapArray[Key, Val]{
		entries: make(map[Key][]Val),
		lock:    sync.RWMutex{},
	}
}
