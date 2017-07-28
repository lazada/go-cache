package cache

import (
	"time"
)

// IStructCache defines required interface for caching module (to be able to store any kind of data, mostly in memory)
type IStructCache interface {
	IFlushable
	RegisterCacheSet(setName string, limit int, ticker *time.Ticker) error
	Get(key *Key) (data interface{}, ok bool)
	GetWithTime(key *Key) (data interface{}, dt time.Time, ok bool)
	Put(data interface{}, key *Key, ttl time.Duration) error
	Remove(key *Key)
	Count() int
	Close()
}

// IStructCacheDebug defines interface for console debug tools
type IStructCacheDebug interface {
	IFlushable
	Count() int
	Find(maskedKey string, limit int) []string
}

type IStructCacheLogger interface {
	IsDebugEnabled() bool
	Debugf(message string, args ...interface{})
	Debug(...interface{})
	Warningf(message string, args ...interface{})
	Warning(...interface{})
}

type ILimitSetter interface {
	SetLimit(int)
}
