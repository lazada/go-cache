package cache

import "time"

const (
	RetryTimeout = time.Second * 10
)

type wrapperCache struct {
	realCache   IByteCache
	stubCache   IByteCache
	isConnected bool
	doneChan    chan bool
	logger      IAerospikeCacheLogger
}

type fnCreate func() (IByteCache, error)

// NewEntryCacheWrapper initializes instance of IEntryCache
func NewEntryCacheWrapper(fn fnCreate, logger IAerospikeCacheLogger) IByteCache {
	result := &wrapperCache{
		stubCache: NewBlackholeCache(),
		logger:    logger,
		doneChan:  make(chan bool, 1),
	}

	go result.createRealCache(fn)

	return result
}

func (this *wrapperCache) getCache() IByteCache {
	if this.isConnected {
		return this.realCache
	}

	return this.stubCache
}

func (this *wrapperCache) createRealCache(fn fnCreate) {
	for {
		select {
		case <-this.doneChan:
			close(this.doneChan)
			return
		case <-time.After(RetryTimeout):
			if cache, err := fn(); err == nil {
				this.realCache = cache
				this.isConnected = true
				this.logger.Debugf("Wrapped cache was created")
				return
			}
		}
	}
}

// Get returns nil, do nothing
func (this *wrapperCache) Get(key *Key) ([]byte, bool) {
	return this.getCache().Get(key)
}

// Put returns nil, do nothing
func (this *wrapperCache) Put(data []byte, key *Key, ttl time.Duration) {
	this.getCache().Put(data, key, ttl)
}

// ScanKeys returns nil, do nothing
func (this *wrapperCache) ScanKeys(set string) ([]Key, error) {
	return this.getCache().ScanKeys(set)
}

// Remove returns nil, do nothing
func (this *wrapperCache) Remove(key *Key) error {
	return this.getCache().Remove(key)
}

// Close do nothing
func (this *wrapperCache) Close() {
	this.doneChan <- true
	this.getCache().Close()
}

// Flush removes all entries from cache and returns number of flushed entries
func (this *wrapperCache) Flush() int {
	return this.getCache().Flush()
}

// Count returns count of data in cache
func (this *wrapperCache) Count() int {
	return this.getCache().Count()
}

// ClearSet returns nil, does nothing
func (this *wrapperCache) ClearSet(set string) error {
	return this.getCache().ClearSet(set)
}
