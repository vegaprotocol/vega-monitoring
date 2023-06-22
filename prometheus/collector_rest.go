package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type RESTCollector struct {
	desc struct {
		coreBlockHeight     *prometheus.Desc
		dataNodeBlockHeight *prometheus.Desc
		coreTime            *prometheus.Desc
		dataNodeTime        *prometheus.Desc
		duration            *prometheus.Desc
		nodeData            *prometheus.Desc
	}
	results map[string]*RESTResults
}

func NewRESTCollector() *RESTCollector {
	collector := &RESTCollector{
		results: map[string]*RESTResults{},
	}

	collector.desc.coreBlockHeight = prometheus.NewDesc(
		"core_block_height", "Current Block Height of Core", []string{"node"}, nil,
	)
	collector.desc.dataNodeBlockHeight = prometheus.NewDesc(
		"data_node_block_height", "Current Block Height of Data-Node", []string{"node"}, nil,
	)
	collector.desc.coreTime = prometheus.NewDesc(
		"core_time", "Current Block Time of Core", []string{"node"}, nil,
	)
	collector.desc.dataNodeTime = prometheus.NewDesc(
		"data_node_time", "Current Block Time of Data-Node", []string{"node"}, nil,
	)
	collector.desc.duration = prometheus.NewDesc(
		"request_rest_statistics_duration", "Response time of /statistics", []string{"node"}, nil,
	)
	collector.desc.nodeData = prometheus.NewDesc(
		"node_data", "Basic information about node", []string{"node", "chain_id", "app_version", "app_version_hash"}, nil,
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
		fieldToValue := map[*prometheus.Desc]float64{
			c.desc.coreBlockHeight:     float64(restResult.CoreBlockHeight),
			c.desc.dataNodeBlockHeight: float64(restResult.DataNodeBlockHeight),
			c.desc.coreTime:            float64(restResult.CoreTime.Unix()),
			c.desc.dataNodeTime:        float64(restResult.DataNodeTime.Unix()),
			c.desc.duration:            restResult.Duration.Seconds(),
		}

		for field, value := range fieldToValue {
			ch <- prometheus.NewMetricWithTimestamp(
				restResult.CurrentTime,
				prometheus.MustNewConstMetric(
					field, prometheus.UntypedValue, value, node,
				))
		}
		ch <- prometheus.NewMetricWithTimestamp(
			restResult.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.nodeData, prometheus.UntypedValue, 1,
				// extra labels
				node, restResult.ChainId, restResult.AppVersion, restResult.AppVersionHash,
			))
	}
}

func (c *RESTCollector) UpdateNodeResults(node string, result *RESTResults) {
	c.results[node] = result
}

func (c *RESTCollector) UpdateNodeAsError(node string, err error) {
	delete(c.results, node)
}
