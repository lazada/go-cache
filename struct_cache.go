package cache

import (
	"strings"
	"sync"
	"time"

	"container/list"

	"go-cache/errors"
	"go-cache/metric"
)

var ErrSetAlreadyExists = errors.New("Set already exists")

type cacheSet struct {
	elements  map[string]*list.Element
	lruList   *list.List
	keysLock  sync.RWMutex
	keysLimit int
	name      string

	logger IStructCacheLogger

	ticker *time.Ticker
	metric metric.Metric

	// connection count metric
	quitCollectorChan chan struct{}
}

// StructCache is simple storage with locking
type StructCache struct {
	setsCollection map[string]*cacheSet
	setsLock       sync.RWMutex
	defaultLimit   int

	logger IStructCacheLogger

	ticker *time.Ticker
	metric metric.Metric
}

var _ IStructCache = &StructCache{} // StructCache implements IStructCache

// NewStructCacheObject returns new instance of StructCache
func NewStructCacheObject(limit int, logger IStructCacheLogger, metric metric.Metric) *StructCache {
	if logger == nil {
		logger = NewNilLogger()
	}

	cache := &StructCache{
		defaultLimit:   limit,
		ticker:         time.NewTicker(5 * time.Minute), // @todo make it changeable param
		setsCollection: make(map[string]*cacheSet),
		logger:         logger,
		metric:         metric,
	}

	if cache.logger.IsDebugEnabled() {
		cache.logger.Debugf("struct_cache: created with %d limit of entries", limit)
	}

	return cache
}

func (cache *StructCache) SetLimit(limit int) {
	if cache.defaultLimit == limit {
		return
	}

	if cache.logger.IsDebugEnabled() {
		cache.logger.Debugf("struct_cache: limit has changed, new value: %d", limit)
	}

	cache.setsLock.Lock()
	cache.defaultLimit = limit
	cache.setsLock.Unlock()
}

func (set *cacheSet) getKeyFromSet(key *Key) (interface{}, time.Time, bool) {

	var (
		created time.Time
		data    interface{}
	)

	set.keysLock.RLock()
	el, ok := set.elements[key.Pk]
	if !ok {
		set.keysLock.RUnlock()
		return data, created, ok
	}

	entry, eok := el.Value.(*Entry)
	set.keysLock.RUnlock()

	if !eok {
		return data, created, false
	}

	set.keysLock.RLock()
	created = entry.CreateDate
	valid := entry.IsValid()
	set.keysLock.RUnlock()

	if !valid {
		set.remove(key)
		return data, created, false
	}

	set.keysLock.RLock()
	data = entry.Data
	set.keysLock.RUnlock()

	set.keysLock.Lock()
	set.lruList.MoveToFront(el)
	set.keysLock.Unlock()

	return data, created, true

}

func (cache *StructCache) getCacheSet(key *Key) (*cacheSet, bool) {
	cache.setsLock.RLock()
	set, exists := cache.setsCollection[key.Set]
	cache.setsLock.RUnlock()

	return set, exists
}

// GetWithTime returns value and create time(UTC) by key
func (cache *StructCache) GetWithTime(key *Key) (interface{}, time.Time, bool) {
	var (
		created time.Time
		data    interface{}
		ok      bool

		ts = time.Now()
	)

	set, setFound := cache.getCacheSet(key)

	if setFound {
		data, created, ok = set.getKeyFromSet(key)
	}

	cache.updateHitOrMissCount(ok, key)

	cache.metric.ObserveRT(map[string]string{
		metric.LabelSet:       key.Set,
		metric.LabelOperation: "get",
	}, metric.SinceMs(ts))

	return data, created, ok
}

func (cache *StructCache) updateHitOrMissCount(condition bool, key *Key) {
	switch condition {
	case true:
		cache.metric.RegisterHit(map[string]string{metric.LabelSet: key.Set})
		if cache.logger.IsDebugEnabled() {
			cache.logger.Debugf("struct_cache: HIT %v", key)
		}
	case false:
		cache.metric.RegisterMiss(map[string]string{metric.LabelSet: key.Set})
		if cache.logger.IsDebugEnabled() {
			cache.logger.Debugf("struct_cache: MISS %v", key)
		}
	}
}

// Get returns value by key
func (cache *StructCache) Get(key *Key) (interface{}, bool) {
	data, _, ok := cache.GetWithTime(key)

	return data, ok
}

// Count elements from cache
func (cache *StructCache) Count() int {
	var count int

	cache.setsLock.RLock()
	for _, set := range cache.setsCollection {
		set.keysLock.RLock()
		count += set.lruList.Len()
		set.keysLock.RUnlock()
	}
	cache.setsLock.RUnlock()

	if cache.logger.IsDebugEnabled() {
		cache.logger.Debugf("struct_cache: count() = %d", count)
	}

	return count
}

// Find search key by mask
func (cache *StructCache) Find(maskedKey string, limit int) []string {
	if cache.logger.IsDebugEnabled() {
		cache.logger.Debugf("struct_cache: FIND %q", maskedKey)
	}

	result := make([]string, 0, limit)

	cache.setsLock.RLock()
	for _, set := range cache.setsCollection {
		set.keysLock.RLock()
		for key := range set.elements {
			if strings.Contains(strings.ToLower(key), maskedKey) {
				result = append(result, key)
				limit--
			}

			if limit <= 0 {
				break
			}
		}
		set.keysLock.RUnlock()

		if limit <= 0 {
			break
		}
	}
	cache.setsLock.RUnlock()

	return result
}

func (cache *StructCache) RegisterCacheSet(setName string, limit int, ticker *time.Ticker) error {
	cache.setsLock.Lock()
	defer cache.setsLock.Unlock()

	if _, exists := cache.setsCollection[setName]; exists {
		return ErrSetAlreadyExists
	}

	cache.setsCollection[setName] = &cacheSet{
		elements:  make(map[string]*list.Element),
		lruList:   list.New(),
		keysLock:  sync.RWMutex{},
		keysLimit: limit,
		name:      setName,

		ticker: ticker,

		logger: cache.logger,
		metric: cache.metric,

		quitCollectorChan: make(chan struct{}, 1),
	}

	if ticker != nil {
		go cache.setsCollection[setName].collector()
	}

	return nil
}

// Put puts elements into storage
func (cache *StructCache) Put(data interface{}, key *Key, ttl time.Duration) error {
	if ttl <= 0 || cache.defaultLimit <= 0 {
		return errors.New("Cannot put element (ttl or cache limit is not assign)")
	}

	if cache.logger.IsDebugEnabled() {
		cache.logger.Debugf("struct_cache: PUT %q with TTL: %s", key, ttl)
	}

	set, exists := cache.getCacheSet(key)
	if !exists {
		cache.RegisterCacheSet(key.Set, cache.defaultLimit, cache.ticker)

		set, exists = cache.getCacheSet(key)
		if !exists {
			return errors.Errorf("Cant create set %q", key.Set)
		}
	}

	return set.put(data, key, ttl)
}

func (set *cacheSet) put(data interface{}, key *Key, ttl time.Duration) error {
	set.keysLock.Lock()
	defer set.keysLock.Unlock()

	ts := time.Now()

	entitiesCount := set.lruList.Len()

	if entitiesCount >= set.keysLimit {
		if set.logger.IsDebugEnabled() {
			set.logger.Debug("struct_cache: ATTENTION! Entities count exceeds limit")
		}
		set.trim()
	}

	if el, ok := set.elements[key.Pk]; ok {
		set.lruList.MoveToFront(el)
		if entry, eok := el.Value.(*Entry); eok {
			entry.EndDate = time.Now().Unix() + int64(ttl.Seconds())

			entry.Data = data
			return nil
		}
	}

	entry := CreateEntry(key, time.Now().Unix()+int64(ttl.Seconds()), data)
	el := set.lruList.PushFront(entry)
	set.elements[key.Pk] = el

	set.metric.ObserveRT(map[string]string{
		metric.LabelSet:       key.Set,
		metric.LabelOperation: "put",
	}, metric.SinceMs(ts))
	set.metric.IncreaseItemCount(set.name)

	return nil
}

func (cache *StructCache) Close() {
	cache.setsLock.RLock()
	defer cache.setsLock.RUnlock()
	for _, set := range cache.setsCollection {
		set.quitCollectorChan <- struct{}{}
	}
}

// trim remove least recently used elements from cache and leave 'limit - 1' elements, to have a change to put one element
func (set *cacheSet) trim() {
	if set.logger.IsDebugEnabled() {
		set.logger.Debugf("struct_cache: trim (max %d current %d)", set.keysLimit, set.lruList.Len())
	}

	for set.lruList.Len() >= set.keysLimit && set.lruList.Len() > 0 {
		el := set.lruList.Back()
		if el != nil {
			if entry, ok := el.Value.(*Entry); ok {
				delete(set.elements, entry.Key.Pk)
				set.lruList.Remove(el)
			}
		}
	}

	set.metric.SetItemCount(set.name, set.lruList.Len())
}

// Remove removes value by key
func (cache *StructCache) Remove(key *Key) {
	set, exists := cache.getCacheSet(key)
	if !exists {
		return
	}

	set.remove(key)
}

func (set *cacheSet) remove(key *Key) {
	k := key.Pk
	if set.logger.IsDebugEnabled() {
		set.logger.Debugf("struct_cache: REMOVE %q", key)
	}

	set.keysLock.Lock()
	if el, ok := set.elements[k]; ok {
		delete(set.elements, k)
		set.lruList.Remove(el)
	}
	set.keysLock.Unlock()

	set.metric.SetItemCount(set.name, set.lruList.Len())
}

func (set *cacheSet) collector() {
	for {
		select {
		case <-set.ticker.C:
			set.keysLock.RLock()
			i := 0
			for _, el := range set.elements {
				if i == 1000 {
					set.keysLock.RUnlock()
					time.Sleep(time.Millisecond * 10)
					i = 0
					set.keysLock.RLock()
				}
				i++
				if entry, ok := el.Value.(*Entry); ok {
					if entry.IsValid() {
						continue
					}
					i--
					if set.logger.IsDebugEnabled() {
						set.logger.Debugf("struct_cache: collector found NOT VALID %q", entry.Key)
					}
					set.keysLock.RUnlock()
					set.remove(entry.Key)
					set.keysLock.RLock()

				}
			}
			set.keysLock.RUnlock()
		case <-set.quitCollectorChan:
			return
		}
	}
}

// Flush removes all entries from cache and returns number of flushed entries
func (cache *StructCache) Flush() int {
	if cache.logger.IsDebugEnabled() {
		cache.logger.Debug("struct_cache: flush()")
	}

	cache.setsLock.RLock()
	defer cache.setsLock.RUnlock()

	var count int

	for _, set := range cache.setsCollection {
		set.keysLock.Lock()
		count += set.lruList.Len()
		set.elements = make(map[string]*list.Element)
		set.lruList.Init()
		set.keysLock.Unlock()

		cache.metric.SetItemCount(set.name, 0)
	}

	return count
}
