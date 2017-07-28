package cache

import (
	"time"

	"go-cache/errors"
)

// StructCacheDummy implements cache that accept any data but always return null
type StructCacheDummy struct {
	logger IStructCacheLogger
}

var _ IStructCache = &StructCacheDummy{}

// NewStructCacheObjectDummy returns new instance of StructCacheDummy
func NewStructCacheObjectDummy(logger IStructCacheLogger) *StructCacheDummy {
	if logger == nil {
		logger = NewNilLogger()
	}
	return &StructCacheDummy{logger}
}

func (cache *StructCacheDummy) RegisterCacheSet(setName string, limit int, ticker *time.Ticker) error {
	cache.logger.Debugf("struct_storage_dummy: registerCacheSet(), setName: %s, limit: %d", setName, limit)
	return nil
}

// Close do nothing
func (cache *StructCacheDummy) Close() {
	cache.logger.Debug("struct_storage_dummy: close()")
}

// GetWithTime returns nil, do nothing
func (cache *StructCacheDummy) GetWithTime(key *Key) (data interface{}, dt time.Time, ok bool) {
	cache.logger.Debugf("struct_storage_dummy: getWithTime(), key: %s", key.ID())
	return
}

// Get returns nil, do nothing
func (cache *StructCacheDummy) Get(key *Key) (data interface{}, ok bool) {
	cache.logger.Debugf("struct_storage_dummy: get(), key: %s", key.ID())
	return
}

// Put returns nil, do nothing
func (cache *StructCacheDummy) Put(data interface{}, key *Key, ttl time.Duration) error {
	cache.logger.Debugf("struct_storage_dummy: put(), key: %s", key.ID())
	// For the beginning, let's check if it is pointer.
	if isPointer(data) {
		return errors.Errorf("struct_storage: It is prohibited to keep pointers in struct cache. %s.", key.String())
	}
	return nil
}

// Remove returns nil, do nothing
func (cache *StructCacheDummy) Remove(key *Key) {
	cache.logger.Debugf("struct_storage_dummy: remove(), key: %s", key.ID())
}

// Count returns number of cache entries
func (cache *StructCacheDummy) Count() int {
	cache.logger.Debug("struct_storage_dummy: count()")
	return 0
}

// Flush removes all entries from cache and returns number of flushed entries
func (cache *StructCacheDummy) Flush() int {
	cache.logger.Debug("struct_storage_dummy: flush()")
	return 0
}
