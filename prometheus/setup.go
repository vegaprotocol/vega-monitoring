package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	rest *RESTCollector
}

func NewMetrics(promRegistry *prometheus.Registry) *Metrics {

	m := &Metrics{
		rest: NewRESTCollector(),
	}

	promRegistry.MustRegister(m.rest)

	return m
}

func (m *Metrics) UpdateNodeCheckResults(node string, results *DataNodeChecksResults) {
	m.rest.UpdateNodeResults(node, results)
}

func (m *Metrics) UpdateNodeAsError(node string, err error) {
	m.rest.UpdateNodeAsError(node, err)
}
