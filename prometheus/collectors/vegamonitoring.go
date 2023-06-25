package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

type VegaMonitoringCollector struct {
	dataNodeStatuses map[string]*types.DataNodeStatus
	dataNodeDown     map[string]error

	blockExplorerStatuses map[string]*types.BlockExplorerStatus
	blockExplorerDown     map[string]error
}

func NewVegaMonitoringCollector() *VegaMonitoringCollector {
	return &VegaMonitoringCollector{
		dataNodeStatuses:      map[string]*types.DataNodeStatus{},
		dataNodeDown:          map[string]error{},
		blockExplorerStatuses: map[string]*types.BlockExplorerStatus{},
		blockExplorerDown:     map[string]error{},
	}
}

func (c *VegaMonitoringCollector) UpdateDataNodeStatus(node string, result *types.DataNodeStatus) {
	delete(c.dataNodeDown, node)
	c.dataNodeStatuses[node] = result
}

func (c *VegaMonitoringCollector) UpdateDataNodeStatusAsError(node string, err error) {
	delete(c.dataNodeStatuses, node)
	c.dataNodeDown[node] = err
}

func (c *VegaMonitoringCollector) UpdateBlockExplorerStatus(node string, result *types.BlockExplorerStatus) {
	delete(c.blockExplorerDown, node)
	c.blockExplorerStatuses[node] = result
}

func (c *VegaMonitoringCollector) UpdateBlockExplorerStatusAsError(node string, err error) {
	delete(c.blockExplorerStatuses, node)
	c.blockExplorerDown[node] = err
}

// Describe returns all descriptions of the collector.
func (c *VegaMonitoringCollector) Describe(ch chan<- *prometheus.Desc) {
	// Core
	ch <- desc.Core.coreBlockHeight
	ch <- desc.Core.coreTime
	ch <- desc.Core.coreInfo

	// DataNode
	ch <- desc.DataNode.dataNodeBlockHeight
	ch <- desc.DataNode.dataNodeTime
	ch <- desc.DataNode.dataNodePerformanceRESTInfoDuration
	ch <- desc.DataNode.dataNodePerformanceGQLInfoDuration
	ch <- desc.DataNode.dataNodePerformanceGRPCInfoDuration
	ch <- desc.DataNode.dataNodeDown

	// BlockExplorer
	ch <- desc.BlockExplorer.blockExplorerInfo
	ch <- desc.BlockExplorer.blockExplorerDown
}

// Collect returns the current state of all metrics of the collector.
func (c *VegaMonitoringCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectDataNodeStatuses(ch)
	c.collectBlockExplorerStatuses(ch)
}

func (c *VegaMonitoringCollector) collectDataNodeStatuses(ch chan<- prometheus.Metric) {
	for node, status := range c.dataNodeStatuses {
		fieldToValue := map[*prometheus.Desc]float64{
			desc.Core.coreBlockHeight:                         float64(status.CoreBlockHeight),
			desc.DataNode.dataNodeBlockHeight:                 float64(status.DataNodeBlockHeight),
			desc.Core.coreTime:                                float64(status.CoreTime.Unix()),
			desc.DataNode.dataNodeTime:                        float64(status.DataNodeTime.Unix()),
			desc.DataNode.dataNodePerformanceRESTInfoDuration: status.RESTReqDuration.Seconds(),
			desc.DataNode.dataNodePerformanceGQLInfoDuration:  status.GQLReqDuration.Seconds(),
			desc.DataNode.dataNodePerformanceGRPCInfoDuration: status.GRPCReqDuration.Seconds(),
		}

		for field, value := range fieldToValue {
			ch <- prometheus.NewMetricWithTimestamp(
				status.CurrentTime,
				prometheus.MustNewConstMetric(
					field, prometheus.UntypedValue, value, node,
				))
		}
		ch <- prometheus.NewMetricWithTimestamp(
			status.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, status.CoreChainId, status.CoreAppVersion, status.CoreAppVersionHash,
			))
	}

	for node := range c.dataNodeDown {
		ch <- prometheus.MustNewConstMetric(
			desc.DataNode.dataNodeDown, prometheus.UntypedValue, 1, node,
		)
	}
}

func (c *VegaMonitoringCollector) collectBlockExplorerStatuses(ch chan<- prometheus.Metric) {
	for node, status := range c.blockExplorerStatuses {
		// Core Block Height
		ch <- prometheus.NewMetricWithTimestamp(
			status.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreBlockHeight, prometheus.UntypedValue, float64(status.CoreBlockHeight),
				// labels
				node,
			))
		// Core Time
		ch <- prometheus.NewMetricWithTimestamp(
			status.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreTime, prometheus.UntypedValue, float64(status.CoreTime.Unix()),
				// labels
				node,
			))
		// Core Info
		ch <- prometheus.NewMetricWithTimestamp(
			status.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, status.CoreChainId, status.CoreAppVersion, status.CoreAppVersionHash,
			))
		// Block Explorer Info
		ch <- prometheus.NewMetricWithTimestamp(
			status.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.BlockExplorer.blockExplorerInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, status.BlockExplorerVersion, status.BlockExplorerVersionHash,
			))
	}

	for node := range c.blockExplorerDown {
		// Block Explorer Down
		ch <- prometheus.MustNewConstMetric(
			desc.BlockExplorer.blockExplorerDown, prometheus.UntypedValue, 1, node,
		)
	}
}
