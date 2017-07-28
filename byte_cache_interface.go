package cache

import "time"

// IByteCache defines required interface for caching module
type IByteCache interface {
	IFlushable
	Count() int
	Get(key *Key) (data []byte, ok bool)
	Put(data []byte, key *Key, ttl time.Duration)
	Remove(key *Key) error
	Close()
	ClearSet(set string) error
	ScanKeys(set string) ([]Key, error)
}

type IMemoryCacheLogger interface {
	Errorf(message string, args ...interface{})
	Warning(message ...interface{})
}

type IAerospikeCacheLogger interface {
	Printf(format string, v ...interface{})
	Debugf(message string, args ...interface{})
	Errorf(message string, args ...interface{})
	Warningf(message string, args ...interface{})
	Warning(...interface{})
	Criticalf(message string, args ...interface{})
	Critical(...interface{})
}
