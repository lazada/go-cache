package cache

import "time"

// IAutoCache defines required interface for caching module (need for auto cache)
type IAutoCache interface {
	Get(key string) (data interface{}, err error)
	Put(updater func() (interface{}, error), key string, ttl time.Duration) error
	Remove(key string)
}

type IAutoCacheLogger interface {
	Errorf(message string, args ...interface{})
	Criticalf(message string, args ...interface{})
}
