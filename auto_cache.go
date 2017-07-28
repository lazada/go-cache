package cache

import (
	"sync"
	"time"

	"go-cache/errors"
)

// StorageAutoCache implements cache that uses auto cache as storage
type StorageAutoCache struct {
	logger  IAutoCacheLogger
	active  bool
	entries map[string]*EntryAutoCache
	lock    sync.RWMutex
}

// NewStorageAutoCacheObject create new instance of StorageAutoCache
func NewStorageAutoCacheObject(active bool, logger IAutoCacheLogger) *StorageAutoCache {
	if logger == nil {
		logger = NewNilLogger()
	}
	return &StorageAutoCache{
		logger:  logger,
		active:  active,
		entries: map[string]*EntryAutoCache{},
	}
}

// Get returns data by given key
func (storage *StorageAutoCache) Get(key string) (interface{}, error) {
	entry, err := storage.getEntry(key)
	if err != nil {
		return nil, err
	}

	return entry.GetValue()
}

func (storage *StorageAutoCache) getEntry(key string) (*EntryAutoCache, error) {
	var err error

	storage.lock.RLock()
	entry, ok := storage.entries[key]
	storage.lock.RUnlock()
	if !ok {
		err = errors.Errorf("Auto cache key %s nof found", key)
	}

	return entry, err
}

// Remove removes value by key
func (storage *StorageAutoCache) Remove(key string) {
	entry, err := storage.getEntry(key)

	if err == nil {
		entry.Stop()
	}

	storage.lock.Lock()
	delete(storage.entries, key)
	storage.lock.Unlock()
}

// Put puts data into storage
func (storage *StorageAutoCache) Put(updater func() (interface{}, error), key string, ttl time.Duration) error {
	entry := CreateEntryAutoCache(updater, ttl, key, storage.logger)

	if storage.active {
		if err := entry.Start(); err != nil {
			return errors.Errorf("Auto cache updater \"%s\" error: %s", key, err)
		}
	}

	storage.lock.Lock()
	if oldEntry, ok := storage.entries[key]; ok {
		oldEntry.Stop()
	}
	storage.entries[key] = entry
	storage.lock.Unlock()

	return nil
}
