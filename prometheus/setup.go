package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	dataNodeCollector *DataNodeCollector
}

func NewMetrics(promRegistry *prometheus.Registry) *Metrics {

	m := &Metrics{
		dataNodeCollector: NewDataNodeCollector(),
	}

	prometheus.WrapRegistererWithPrefix("vega_monitoring", promRegistry).MustRegister(m.dataNodeCollector)

	return m
}

func (m *Metrics) UpdateDataNodeCheckResults(node string, results *DataNodeChecksResults) {
	m.dataNodeCollector.UpdateNodeResults(node, results)
}

func (m *Metrics) UpdateDataNodeAsError(node string, err error) {
	m.dataNodeCollector.UpdateNodeAsError(node, err)
}
