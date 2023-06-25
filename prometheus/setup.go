package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	dataNodeCollector      *DataNodeCollector
	blockExplorerCollector *BlockExplorerCollector
}

func NewMetrics(promRegistry *prometheus.Registry) *Metrics {

	metricsRegisterer := prometheus.WrapRegistererWithPrefix("vega_monitoring", promRegistry)

	m := &Metrics{
		dataNodeCollector:      NewDataNodeCollector(),
		blockExplorerCollector: NewBlockExplorerCollector(),
	}

	metricsRegisterer.MustRegister(
		m.dataNodeCollector,
		m.blockExplorerCollector,
	)

	return m
}

func (m *Metrics) UpdateDataNodeChecksResults(node string, results *DataNodeChecksResults) {
	m.dataNodeCollector.UpdateNodeResults(node, results)
}

func (m *Metrics) UpdateDataNodeAsError(node string, err error) {
	m.dataNodeCollector.UpdateNodeAsError(node, err)
}

func (m *Metrics) UpdateBlockExplorerChecksResults(node string, results *BlockExplorerChecksResults) {
	m.blockExplorerCollector.UpdateNodeResults(node, results)
}

func (m *Metrics) UpdateBlockExplorerAsError(node string, err error) {
	m.blockExplorerCollector.UpdateNodeAsError(node, err)
}
