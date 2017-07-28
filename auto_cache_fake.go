package cache

import (
	"sync"
	"time"

	"go-cache/errors"
)

// StorageAutoCacheFake implements cache that uses auto cache as storage
// This type of autocache have no TTL
type StorageAutoCacheFake struct {
	updaters map[string]func() (interface{}, error)
	mutex    sync.RWMutex
}

// NewStorageAutoCache create new instance of StorageAutoCache
func NewStorageAutoCacheFake() *StorageAutoCacheFake {
	return &StorageAutoCacheFake{
		updaters: make(map[string]func() (interface{}, error)),
	}
}

// Get returns data by given key
func (storage *StorageAutoCacheFake) Get(key string) (interface{}, error) {
	storage.mutex.RLock()
	updater, find := storage.updaters[key]
	storage.mutex.RUnlock()
	if !find {
		return nil, errors.Errorf("Auto cache key %s nof found", key)
	}

	return updater()
}

// Remove removes value by key
func (storage *StorageAutoCacheFake) Remove(key string) {
	storage.mutex.Lock()
	delete(storage.updaters, key)
	storage.mutex.Unlock()
}

// Put puts data into storage
func (storage *StorageAutoCacheFake) Put(updater func() (interface{}, error), key string, ttl time.Duration) error {
	storage.mutex.Lock()
	storage.updaters[key] = updater
	storage.mutex.Unlock()

	return nil
}
