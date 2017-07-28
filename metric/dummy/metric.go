package dummy

type Metric struct{}

func NewMetric() Metric {
	return Metric{}
}

func (m Metric) ObserveRT(labels map[string]string, timeSince float64) {
	return
}

func (m Metric) RegisterHit(labels map[string]string) {
	return
}

func (m Metric) RegisterMiss(labels map[string]string) {
	return
}

func (m Metric) IncreaseItemCount(set string) {
	return
}

func (m Metric) SetItemCount(set string, n int) {
	return
}
