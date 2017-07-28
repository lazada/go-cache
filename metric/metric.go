package metric

import "time"

const (
	LabelHost      = "host"
	LabelIsError   = "is_error"
	LabelNamespace = "namespace"
	LabelSet       = "set"
	LabelOperation = "operation"
)

type Metric interface {
	ObserveRT(labels map[string]string, timeSince float64)
	RegisterHit(labels map[string]string)
	RegisterMiss(labels map[string]string)
	IncreaseItemCount(set string)
	SetItemCount(set string, n int)
}

// SinceMs just wraps time.Since() with converting result to milliseconds.
// Because Prometheus prefers milliseconds.
func SinceMs(started time.Time) float64 {
	return float64(time.Since(started)) / float64(time.Millisecond)
}

// IsError is a trivial helper for minimize repetitive checks for error
// values. It passing appropriate numbers to metrics.
func IsError(err error) string {
	if err != nil {
		return "1"
	}
	return "0"
}
