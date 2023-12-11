package cache

import "sync"

type RWLockMap[Key comparable, Val any] struct {
	entries map[Key]Val
	lock    sync.RWMutex
}

type RWLockMapArray[Key comparable, Val any] struct {
	entries map[Key][]Val
	lock    sync.RWMutex
}

type RWLockMapMap[Key comparable, KeyInner comparable, Val any] struct {
	entries map[Key]map[KeyInner]Val
	lock    sync.RWMutex
}

func (m *RWLockMap[Key, Val]) lookup(key Key) (Val, bool) {
	m.lock.RLock()
	result, ok := m.entries[key]
	m.lock.RUnlock()
	return result, ok
}

func (m *RWLockMap[Key, Val]) keySet() []Key {
	keys := make([]Key, 0)
	m.lock.RLock()
	for key := range m.entries {
		keys = append(keys, key)
	}
	m.lock.RUnlock()
	return keys
}

func (m *RWLockMap[Key, Val]) clone() map[Key]Val {
	cloned := make(map[Key]Val, 0)
	m.lock.RLock()
	for key, val := range m.entries {
		cloned[key] = val
	}
	m.lock.RUnlock()
	return cloned
}

func (m *RWLockMapArray[Key, Val]) lookup(key Key) ([]Val, bool) {
	m.lock.RLock()
	result, ok := m.entries[key]
	m.lock.RUnlock()
	return result, ok
}

func (m *RWLockMapArray[Key, Val]) lookupCopy(key Key) ([]Val, bool) {
	var retVal []Val
	m.lock.RLock()
	result, ok := m.entries[key]
	if ok {
		for _, r := range result {
			retVal = append(retVal, r)
		}
	}
	m.lock.RUnlock()
	return retVal, ok
}

func (m *RWLockMapMap[Key, KeyInner, Val]) lookupSet(key Key) (map[KeyInner]Val, bool) {
	var retVal map[KeyInner]Val = make(map[KeyInner]Val, 0)
	m.lock.RLock()
	result, ok := m.entries[key]
	if ok {
		for k, v := range result {
			retVal[k] = v
		}
	}
	m.lock.RUnlock()
	return result, ok
}

func (m *RWLockMap[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	m.entries[key] = val
	m.lock.Unlock()
}

func (m *RWLockMapArray[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	result, ok := m.entries[key]
	if !ok {
		m.entries[key] = make([]Val, 0)
	}
	m.entries[key] = append(result, val)
	m.lock.Unlock()
}

func (m *RWLockMapMap[Key, KeyInner, Val]) insert(key Key, keyIn KeyInner, val Val) {
	m.lock.Lock()
	_, ok := m.entries[key]
	if !ok {
		m.entries[key] = make(map[KeyInner]Val, 0)
	}
	m.entries[key][keyIn] = val
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

func newRwLockMapMap[Key comparable, KeyInner comparable, Val any]() RWLockMapMap[Key, KeyInner, Val] {
	return RWLockMapMap[Key, KeyInner, Val]{
		entries: make(map[Key]map[KeyInner]Val),
		lock:    sync.RWMutex{},
	}
}
