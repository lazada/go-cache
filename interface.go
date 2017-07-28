package cache

// IFlushable determines cache that can be flushed
type IFlushable interface {
	Flush() int
}
