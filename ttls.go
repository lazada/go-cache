package cache

import "time"

const (
	// SmallCacheTTL define a short cache lifetime
	SmallCacheTTL = 5 * time.Minute

	// DefaultCacheTTL default cache lifetime
	DefaultCacheTTL = 30 * time.Minute

	// TwoHCacheTTL two hours cache lifetime
	TwoHCacheTTL = 2 * time.Hour

	// LongCacheTTL large cache lifetime (one day)
	LongCacheTTL = 24 * time.Hour

	// VeryLongCacheTTL very large cache lifetime (one mounth)
	VeryLongCacheTTL = 30 * 24 * time.Hour

	// EternalLongCacheTTL eternal large cache lifetime (one 10 years)
	EternalLongCacheTTL = 10 * 365 * 24 * time.Hour
)
