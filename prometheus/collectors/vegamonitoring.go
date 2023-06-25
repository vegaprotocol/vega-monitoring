package collectors

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

type VegaMonitoringCollector struct {
	dataNodeStatuses map[string]*types.DataNodeStatus
	dataNodeDown     map[string]types.NodeDownStatus

	blockExplorerStatuses map[string]*types.BlockExplorerStatus
	blockExplorerDown     map[string]types.NodeDownStatus
}

func NewVegaMonitoringCollector() *VegaMonitoringCollector {
	return &VegaMonitoringCollector{
		dataNodeStatuses:      map[string]*types.DataNodeStatus{},
		dataNodeDown:          map[string]types.NodeDownStatus{},
		blockExplorerStatuses: map[string]*types.BlockExplorerStatus{},
		blockExplorerDown:     map[string]types.NodeDownStatus{},
	}
}

func (c *VegaMonitoringCollector) UpdateDataNodeStatus(node string, result *types.DataNodeStatus) {
	delete(c.dataNodeDown, node)
	c.dataNodeStatuses[node] = result
}

func (c *VegaMonitoringCollector) UpdateDataNodeStatusAsError(node string, downStatus types.NodeDownStatus) {
	delete(c.dataNodeStatuses, node)
	c.dataNodeDown[node] = downStatus
}

func (c *VegaMonitoringCollector) UpdateBlockExplorerStatus(node string, result *types.BlockExplorerStatus) {
	delete(c.blockExplorerDown, node)
	c.blockExplorerStatuses[node] = result
}

func (c *VegaMonitoringCollector) UpdateBlockExplorerStatusAsError(node string, downStatus types.NodeDownStatus) {
	delete(c.blockExplorerStatuses, node)
	c.blockExplorerDown[node] = downStatus
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
	for nodeName, nodeStatus := range c.dataNodeStatuses {
		fieldToValue := map[*prometheus.Desc]float64{
			desc.Core.coreBlockHeight:                         float64(nodeStatus.CoreBlockHeight),
			desc.DataNode.dataNodeBlockHeight:                 float64(nodeStatus.DataNodeBlockHeight),
			desc.Core.coreTime:                                float64(nodeStatus.CoreTime.Unix()),
			desc.DataNode.dataNodeTime:                        float64(nodeStatus.DataNodeTime.Unix()),
			desc.DataNode.dataNodePerformanceRESTInfoDuration: nodeStatus.RESTReqDuration.Seconds(),
			desc.DataNode.dataNodePerformanceGQLInfoDuration:  nodeStatus.GQLReqDuration.Seconds(),
			desc.DataNode.dataNodePerformanceGRPCInfoDuration: nodeStatus.GRPCReqDuration.Seconds(),
		}

		for field, value := range fieldToValue {
			ch <- prometheus.NewMetricWithTimestamp(
				nodeStatus.CurrentTime,
				prometheus.MustNewConstMetric(
					field, prometheus.UntypedValue, value,
					// Labels
					nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
				))
		}
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreInfo, prometheus.UntypedValue, 1,
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
				// Extra labels
				nodeStatus.CoreChainId, nodeStatus.CoreAppVersion, nodeStatus.CoreAppVersionHash,
			))
	}

	for nodeName, nodeStatus := range c.dataNodeDown {
		ch <- prometheus.MustNewConstMetric(
			desc.DataNode.dataNodeDown, prometheus.UntypedValue, 1,
			// Labels
			nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
		)
	}
}

func (c *VegaMonitoringCollector) collectBlockExplorerStatuses(ch chan<- prometheus.Metric) {
	for nodeName, nodeStatus := range c.blockExplorerStatuses {
		// Core Block Height
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreBlockHeight, prometheus.UntypedValue, float64(nodeStatus.CoreBlockHeight),
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
			))
		// Core Time
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreTime, prometheus.UntypedValue, float64(nodeStatus.CoreTime.Unix()),
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
			))
		// Core Info
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.Core.coreInfo, prometheus.UntypedValue, 1,
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
				// Extra labels
				nodeStatus.CoreChainId, nodeStatus.CoreAppVersion, nodeStatus.CoreAppVersionHash,
			))
		// Block Explorer Info
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.BlockExplorer.blockExplorerInfo, prometheus.UntypedValue, 1,
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
				// Extra labels
				nodeStatus.BlockExplorerVersion, nodeStatus.BlockExplorerVersionHash,
			))
	}

	for nodeName, nodeStatus := range c.blockExplorerDown {
		// Block Explorer Down
		ch <- prometheus.MustNewConstMetric(
			desc.BlockExplorer.blockExplorerDown, prometheus.UntypedValue, 1,
			// Labels
			nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
		)
	}
}
