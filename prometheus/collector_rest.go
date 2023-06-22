package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type RESTCollector struct {
	desc struct {
		coreBlockHeight *prometheus.Desc
	}
	results map[string]*RESTResults
}

func NewRESTCollector() *RESTCollector {
	collector := &RESTCollector{
		results: map[string]*RESTResults{},
	}

	collector.desc.coreBlockHeight = prometheus.NewDesc(
		"core_block_height", "Current Block Height", []string{"node"}, nil,
	)

	return collector
}

// Describe returns all descriptions of the collector.
func (c *RESTCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc.coreBlockHeight
}

// Collect returns the current state of all metrics of the collector.
func (c *RESTCollector) Collect(ch chan<- prometheus.Metric) {
	for node, restResult := range c.results {
		ch <- prometheus.NewMetricWithTimestamp(
			restResult.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.coreBlockHeight, prometheus.UntypedValue, float64(restResult.CoreBlockHeight), node,
			))
	}
}

func (c *RESTCollector) UpdateNodeResults(node string, result *RESTResults) {
	c.results[node] = result
}

func (c *RESTCollector) UpdateNodeAsError(node string, err error) {
	delete(c.results, node)
}
