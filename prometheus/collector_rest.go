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
		nodeData            *prometheus.Desc

		restReqDuration *prometheus.Desc
		gqlReqDuration  *prometheus.Desc
		grpcReqDuration *prometheus.Desc

		dataNodeDown *prometheus.Desc
	}
	results      map[string]*DataNodeChecksResults
	dataNodeDown map[string]error
}

func NewRESTCollector() *RESTCollector {
	collector := &RESTCollector{
		results:      map[string]*DataNodeChecksResults{},
		dataNodeDown: map[string]error{},
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
	collector.desc.nodeData = prometheus.NewDesc(
		"node_data", "Basic information about node", []string{"node", "chain_id", "app_version", "app_version_hash"}, nil,
	)

	collector.desc.restReqDuration = prometheus.NewDesc(
		"request_rest_duration", "Duration of REST request to get info about node", []string{"node"}, nil,
	)
	collector.desc.gqlReqDuration = prometheus.NewDesc(
		"request_gql_duration", "Duration of GraphQL request to get info about node", []string{"node"}, nil,
	)
	collector.desc.grpcReqDuration = prometheus.NewDesc(
		"request_grpc_duration", "Duration of gRPC request to get info about node", []string{"node"}, nil,
	)

	collector.desc.dataNodeDown = prometheus.NewDesc(
		"data_node_down", "Data Node is not responsive", []string{"node"}, nil,
	)

	return collector
}

// Describe returns all descriptions of the collector.
func (c *RESTCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc.coreBlockHeight
	ch <- c.desc.dataNodeBlockHeight
	ch <- c.desc.coreTime
	ch <- c.desc.dataNodeTime
	ch <- c.desc.nodeData
	ch <- c.desc.restReqDuration
	ch <- c.desc.gqlReqDuration
	ch <- c.desc.grpcReqDuration
	ch <- c.desc.dataNodeDown
}

// Collect returns the current state of all metrics of the collector.
func (c *RESTCollector) Collect(ch chan<- prometheus.Metric) {
	for node, checkResults := range c.results {
		fieldToValue := map[*prometheus.Desc]float64{
			c.desc.coreBlockHeight:     float64(checkResults.CoreBlockHeight),
			c.desc.dataNodeBlockHeight: float64(checkResults.DataNodeBlockHeight),
			c.desc.coreTime:            float64(checkResults.CoreTime.Unix()),
			c.desc.dataNodeTime:        float64(checkResults.DataNodeTime.Unix()),
			c.desc.restReqDuration:     checkResults.RESTReqDuration.Seconds(),
			c.desc.gqlReqDuration:      checkResults.GQLReqDuration.Seconds(),
			c.desc.grpcReqDuration:     checkResults.GRPCReqDuration.Seconds(),
		}

		for field, value := range fieldToValue {
			ch <- prometheus.NewMetricWithTimestamp(
				checkResults.CurrentTime,
				prometheus.MustNewConstMetric(
					field, prometheus.UntypedValue, value, node,
				))
		}
		ch <- prometheus.NewMetricWithTimestamp(
			checkResults.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.nodeData, prometheus.UntypedValue, 1,
				// extra labels
				node, checkResults.ChainId, checkResults.AppVersion, checkResults.AppVersionHash,
			))
	}

	for node, _ := range c.dataNodeDown {
		ch <- prometheus.MustNewConstMetric(
			c.desc.dataNodeDown, prometheus.UntypedValue, 1, node,
		)
	}
}

func (c *RESTCollector) UpdateNodeResults(node string, result *DataNodeChecksResults) {
	delete(c.dataNodeDown, node)
	c.results[node] = result
}

func (c *RESTCollector) UpdateNodeAsError(node string, err error) {
	delete(c.results, node)
	c.dataNodeDown[node] = err
}
