package cache

import (
	"bytes"
	"slices"
	"sync"
)

type RWLockMap[Key comparable, Val any] struct {
	entries    map[Key]Val
	entryLocks map[Key]*sync.RWMutex
	lock       sync.RWMutex
}

// HasTime constraint is necessary only for lookupLastNTime, otherwise it
// would require type assertion for every element, which is very slow.
type RWLockMapArray[Key comparable, Val interface {
	HasTime
	HasHash
}] struct {
	entries map[Key][]Val
	lock    sync.RWMutex
}

// HasTime constraint is necessary only for lookupLastNTime, otherwise it
// would require type assertion for every element, which is very slow.
type RWLockMapMap[Key comparable, KeyInner interface {
	comparable
	HasHash
}, Val HasTime] struct {
	entries map[Key]map[KeyInner]Val
	lock    sync.RWMutex
}

func (m *RWLockMap[Key, Val]) lookup(key Key) (result Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result, ok = m.entries[key]
	return
}

func (m *RWLockMap[Key, Val]) lockLookup(key Key, writeLock bool) (result Val, ok bool, lock *sync.RWMutex) {
	m.lock.RLock()
	result, ok = m.entries[key]
	if ok {
		lock = m.entryLocks[key]
		if writeLock {
			lock.Lock()
		} else {
			lock.RLock()
		}
	}
	m.lock.RUnlock()
	return
}

func (m *RWLockMap[Key, Val]) keySet() []Key {
	keys := make([]Key, 0)
	m.lock.RLock()
	defer m.lock.RUnlock()
	for key := range m.entries {
		keys = append(keys, key)
	}
	return keys
}

func (m *RWLockMap[Key, Val]) clone() map[Key]Val {
	cloned := make(map[Key]Val, 0)
	m.lock.RLock()
	defer m.lock.RUnlock()
	for key, val := range m.entries {
		cloned[key] = val
	}
	return cloned
}

func (m *RWLockMapArray[Key, Val]) lookup(key Key) (result []Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result, ok = m.entries[key]
	return
}

func (m *RWLockMapArray[Key, Val]) lookupCopy(key Key) (retVal []Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	rows, ok := m.entries[key]
	if ok {
		retVal = append(retVal, rows...)
	}
	return
}

func (m *RWLockMapArray[Key, Val]) lookupLastN(key Key, lastN int) (retVal []Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	rows, ok := m.entries[key]
	if ok {
		if len(rows) < lastN {
			lastN = len(rows)
		}

		retVal = append(retVal, rows[len(rows)-lastN:]...)
	}
	slices.Reverse(retVal)
	return
}

// Fast lookup for last N elements in time range. It assumes that the array
// is sorted by time and all elements are unique (as is the case for userTxs/poolTxs).
func (m *RWLockMapArray[Key, Val]) lookupLastNAtTime(key Key, afterTime int, beforeTime int, n int) (result []Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	rows, ok := m.entries[key]
	if ok {
		result = make([]Val, 0, n)
		for i := len(rows) - 1; i >= 0; i-- {
			t := rows[i].Time()
			if t >= afterTime && t < beforeTime {
				result = append(result, rows[i])
				if len(result) >= n {
					break
				}
			}
			if t < afterTime {
				break
			}
		}
	}
	return
}

// Version of lookupLastNAtTime that's used for poolPosUpdates and poolKoUpdates because
// they have entries for updates so the same position/order will be stored multiple times.
// `seen` is passed in to not reallocate it every time for subsequent calls.
func (m *RWLockMapArray[Key, Val]) lookupLastNTimeNonUnique(key Key, afterTime int, beforeTime int, n int, seen map[[32]byte]struct{}) (result []Val, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	rows, ok := m.entries[key]
	if ok {
		result = make([]Val, 0, n)
		buf := new(bytes.Buffer)
		buf.Grow(300)
		for i := len(rows) - 1; i >= 0; i-- {
			hash := rows[i].Hash(buf)
			if _, ok := seen[hash]; ok {
				continue
			}
			seen[hash] = struct{}{}
			t := rows[i].Time()
			if t >= afterTime && t < beforeTime {
				result = append(result, rows[i])
				if len(result) >= n {
					break
				}
			}
			if t < afterTime {
				break
			}
		}
	}
	return
}

type HasTime interface {
	Time() int
}

type HasHash interface {
	// buf is optional to avoid unnecessary allocations
	Hash(buf *bytes.Buffer) [32]byte
}

// Note this function locks the map and returns a reference to map entry. It's very important
// for the caller to unlock when complete, otherwise the map will be locked indefinitely.
func (m *RWLockMapMap[Key, KeyInner, Val]) lockSet(key Key) (result map[KeyInner]Val, lock *sync.RWMutex) {
	m.lock.RLock()
	result, ok := m.entries[key]
	if !ok {
		m.lock.RUnlock()
		return make(map[KeyInner]Val, 0), nil
	}
	return result, &m.lock
}

func (m *RWLockMapMap[Key, KeyInner, Val]) lookupSet(key Key) (retVal map[KeyInner]Val, ok bool) {
	retVal = make(map[KeyInner]Val, 0)
	m.lock.RLock()
	defer m.lock.RUnlock()
	entries, ok := m.entries[key]
	if ok {
		for k, v := range entries {
			retVal[k] = v
		}
	}
	return
}

func (m *RWLockMap[Key, Val]) insert(key Key, val Val) *sync.RWMutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.entries[key] = val
	m.entryLocks[key] = &sync.RWMutex{}
	return m.entryLocks[key]
}

func (m *RWLockMapArray[Key, Val]) insert(key Key, val Val) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.entries[key] = append(m.entries[key], val)
}

func (m *RWLockMapArray[Key, Val]) insertSorted(key Key, val Val, less func(i, j Val) bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
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
}

func (m *RWLockMapMap[Key, KeyInner, Val]) insert(key Key, keyIn KeyInner, val Val) {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.entries[key]
	if !ok {
		m.entries[key] = make(map[KeyInner]Val, 0)
	}
	m.entries[key][keyIn] = val
}

func newRwLockMap[Key comparable, Val any]() RWLockMap[Key, Val] {
	return RWLockMap[Key, Val]{
		entries:    make(map[Key]Val),
		lock:       sync.RWMutex{},
		entryLocks: make(map[Key]*sync.RWMutex),
	}
}

func newRwLockMapArray[Key comparable, Val interface {
	HasTime
	HasHash
}]() RWLockMapArray[Key, Val] {
	return RWLockMapArray[Key, Val]{
		entries: make(map[Key][]Val),
		lock:    sync.RWMutex{},
	}
}

func newRwLockMapMap[Key comparable, KeyInner interface {
	comparable
	HasHash
}, Val HasTime]() RWLockMapMap[Key, KeyInner, Val] {
	return RWLockMapMap[Key, KeyInner, Val]{
		entries: make(map[Key]map[KeyInner]Val),
		lock:    sync.RWMutex{},
	}
}
