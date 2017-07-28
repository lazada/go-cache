package cache

import "time"

// BlackholeCache implements cache that accept any data but always return null. Used for unit tests
type BlackholeCache struct {
}

var _ IByteCache = &BlackholeCache{} // BlackholeCache implements IByteCache

// NewBlackholeCache initializes instance of BlackholeCache
func NewBlackholeCache() *BlackholeCache {
	return &BlackholeCache{}
}

// Get returns nil, do nothing
func (cache *BlackholeCache) Get(key *Key) (data []byte, ok bool) {
	return
}

// Put returns nil, do nothing
func (cache *BlackholeCache) Put(data []byte, key *Key, ttl time.Duration) {}

// ScanKeys returns nil, do nothing
func (cache *BlackholeCache) ScanKeys(set string) ([]Key, error) {
	return nil, nil
}

// Remove returns nil, do nothing
func (cache *BlackholeCache) Remove(key *Key) (err error) {
	return
}

// Close do nothing
func (cache *BlackholeCache) Close() {}

// Flush removes all entries from cache and returns number of flushed entries
func (cache *BlackholeCache) Flush() int {
	return 0
}

// Count returns count of data in cache FIXME: implement me
func (cache *BlackholeCache) Count() int {
	return 0
}

// ClearSet returns nil, does nothing
func (cache *BlackholeCache) ClearSet(set string) error {
	return nil
}
