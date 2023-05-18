package model

import (
	"sync"
	"time"
)

type ExpiryHandle[T any] struct {
	data         T
	fetchFn      func() T
	expireTime   int64
	expirePeriod int64
	wg           sync.WaitGroup
	lock         sync.RWMutex
}

func InitCacheHandle[T any](fetchFn func() T, expirePeriod int64, initVal T) *ExpiryHandle[T] {
	hndl := ExpiryHandle[T]{
		data:         initVal,
		fetchFn:      fetchFn,
		expirePeriod: expirePeriod,
	}

	hndl.wg.Add(1)
	go hndl.fetchFn()
	return &hndl
}

func (hndl *ExpiryHandle[T]) Poll() T {
	if hndl.expireTime > time.Now().Unix() {
		hndl.wg.Wait()
	}

	hndl.lock.RLock()
	defer hndl.lock.RUnlock()
	return hndl.data
}

func (hndl *ExpiryHandle[T]) refresh() {
	hndl.wg.Add(1)

	go func() {
		hndl.lock.Lock()
		hndl.data = hndl.fetchFn()
		hndl.lock.Unlock()

		hndl.expireTime = time.Now().Unix() + hndl.expirePeriod
		hndl.wg.Done()
	}()
}
