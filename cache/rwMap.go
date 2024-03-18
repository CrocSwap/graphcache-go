package cache

import (
	"slices"
	"sync"
)

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
		retVal = append(retVal, result...)
	}
	m.lock.RUnlock()
	return retVal, ok
}

func (m *RWLockMapArray[Key, Val]) lookupLastN(key Key, lastN int) ([]Val, bool) {
	var retVal []Val
	m.lock.RLock()
	result, ok := m.entries[key]
	if ok {
		if len(result) < lastN {
			lastN = len(result)
		}

		retVal = append(retVal, result[len(result)-lastN:]...)
	}
	m.lock.RUnlock()
	slices.Reverse(retVal)
	return retVal, ok
}

// Note this function locks the map and returns a reference to map entry. It's very important
// for the caller to unlock when complete, otherwise the map will be locked indefinitely.
func (m *RWLockMapMap[Key, KeyInner, Val]) lockSet(key Key) (map[KeyInner]Val, *sync.RWMutex) {
	m.lock.RLock()
	result, ok := m.entries[key]
	if !ok {
		m.lock.RUnlock()
		return make(map[KeyInner]Val, 0), nil
	}
	return result, &m.lock
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
	return retVal, ok
}

func (m *RWLockMap[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	m.entries[key] = val
	m.lock.Unlock()
}

func (m *RWLockMapArray[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	m.entries[key] = append(m.entries[key], val)
	m.lock.Unlock()
}

func (m *RWLockMapArray[Key, Val]) insertSorted(key Key, val Val, less func(i, j Val) bool) {
	m.lock.Lock()
	result := m.entries[key]
	var i int
	for i = len(result) - 1; i >= 0; i-- {
		if less(val, result[i]) {
			break
		}
	}
	i += 1

	var zero Val
	result = append(result, zero)
	copy(result[i+1:], result[i:])
	result[i] = val
	m.entries[key] = result
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
