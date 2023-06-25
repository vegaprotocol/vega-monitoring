package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type DataNodeCollector struct {
	desc struct {
		coreBlockHeight     *prometheus.Desc
		dataNodeBlockHeight *prometheus.Desc
		coreTime            *prometheus.Desc
		dataNodeTime        *prometheus.Desc
		coreInfo            *prometheus.Desc

		dataNodePerformanceRESTInfoDuration *prometheus.Desc
		dataNodePerformanceGQLInfoDuration  *prometheus.Desc
		dataNodePerformanceGRPCInfoDuration *prometheus.Desc

		dataNodeDown *prometheus.Desc
	}
	results      map[string]*DataNodeChecksResults
	dataNodeDown map[string]error
}

func NewDataNodeCollector() *DataNodeCollector {
	collector := &DataNodeCollector{
		results:      map[string]*DataNodeChecksResults{},
		dataNodeDown: map[string]error{},
	}

	collector.desc.coreBlockHeight = prometheus.NewDesc(
		"core_block_height", "Current Block Height of Core", []string{"node"}, nil,
	)
	collector.desc.dataNodeBlockHeight = prometheus.NewDesc(
		"datanode_block_height", "Current Block Height of Data-Node", []string{"node"}, nil,
	)
	collector.desc.coreTime = prometheus.NewDesc(
		"core_time", "Current Block Time of Core", []string{"node"}, nil,
	)
	collector.desc.dataNodeTime = prometheus.NewDesc(
		"datanode_time", "Current Block Time of Data-Node", []string{"node"}, nil,
	)
	collector.desc.coreInfo = prometheus.NewDesc(
		"core_info", "Basic information about node", []string{"node", "chain_id", "app_version", "app_version_hash"}, nil,
	)

	collector.desc.dataNodePerformanceRESTInfoDuration = prometheus.NewDesc(
		"datanode_performance_rest_info_duration", "Duration of REST request to get info about node", []string{"node"}, nil,
	)
	collector.desc.dataNodePerformanceGQLInfoDuration = prometheus.NewDesc(
		"datanode_performance_gql_info_duration", "Duration of GraphQL request to get info about node", []string{"node"}, nil,
	)
	collector.desc.dataNodePerformanceGRPCInfoDuration = prometheus.NewDesc(
		"datanode_performance_grpc_info_duration", "Duration of gRPC request to get info about node", []string{"node"}, nil,
	)

	collector.desc.dataNodeDown = prometheus.NewDesc(
		"datanode_down", "Data Node is not responsive", []string{"node"}, nil,
	)

	return collector
}

// Describe returns all descriptions of the collector.
func (c *DataNodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc.coreBlockHeight
	ch <- c.desc.dataNodeBlockHeight
	ch <- c.desc.coreTime
	ch <- c.desc.dataNodeTime
	ch <- c.desc.coreInfo
	ch <- c.desc.dataNodePerformanceRESTInfoDuration
	ch <- c.desc.dataNodePerformanceGQLInfoDuration
	ch <- c.desc.dataNodePerformanceGRPCInfoDuration
	ch <- c.desc.dataNodeDown
}

// Collect returns the current state of all metrics of the collector.
func (c *DataNodeCollector) Collect(ch chan<- prometheus.Metric) {
	for node, checkResults := range c.results {
		fieldToValue := map[*prometheus.Desc]float64{
			c.desc.coreBlockHeight:                     float64(checkResults.CoreBlockHeight),
			c.desc.dataNodeBlockHeight:                 float64(checkResults.DataNodeBlockHeight),
			c.desc.coreTime:                            float64(checkResults.CoreTime.Unix()),
			c.desc.dataNodeTime:                        float64(checkResults.DataNodeTime.Unix()),
			c.desc.dataNodePerformanceRESTInfoDuration: checkResults.RESTReqDuration.Seconds(),
			c.desc.dataNodePerformanceGQLInfoDuration:  checkResults.GQLReqDuration.Seconds(),
			c.desc.dataNodePerformanceGRPCInfoDuration: checkResults.GRPCReqDuration.Seconds(),
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
				c.desc.coreInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, checkResults.CoreChainId, checkResults.CoreAppVersion, checkResults.CoreAppVersionHash,
			))
	}

	for node, _ := range c.dataNodeDown {
		ch <- prometheus.MustNewConstMetric(
			c.desc.dataNodeDown, prometheus.UntypedValue, 1, node,
		)
	}
}

func (c *DataNodeCollector) UpdateNodeResults(node string, result *DataNodeChecksResults) {
	delete(c.dataNodeDown, node)
	c.results[node] = result
}

func (c *DataNodeCollector) UpdateNodeAsError(node string, err error) {
	delete(c.results, node)
	c.dataNodeDown[node] = err
}
