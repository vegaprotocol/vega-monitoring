package collectors

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
	"github.com/vegaprotocol/vega-monitoring/services/read"
)

type VegaMonitoringCollector struct {
	coreStatuses          map[string]*types.CoreStatus
	dataNodeStatuses      map[string]*types.DataNodeStatus
	blockExplorerStatuses map[string]*types.BlockExplorerStatus
	nodeDownStatuses      map[string]types.NodeDownStatus

	// Meta-Monitoring
	monitoringDatabaseStatuses read.MetaMonitoringStatuses

	accessMu sync.RWMutex
}

func NewVegaMonitoringCollector() *VegaMonitoringCollector {
	return &VegaMonitoringCollector{
		coreStatuses:          map[string]*types.CoreStatus{},
		dataNodeStatuses:      map[string]*types.DataNodeStatus{},
		blockExplorerStatuses: map[string]*types.BlockExplorerStatus{},
		nodeDownStatuses:      map[string]types.NodeDownStatus{},
	}
}

func (c *VegaMonitoringCollector) UpdateCoreStatus(node string, newStatus *types.CoreStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.coreStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) UpdateDataNodeStatus(node string, newStatus *types.DataNodeStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.dataNodeStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) UpdateBlockExplorerStatus(node string, newStatus *types.BlockExplorerStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.blockExplorerStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) UpdateNodeStatusAsError(node string, newStatus types.NodeDownStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.nodeDownStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) clearStatusFor(node string) {
	delete(c.coreStatuses, node)
	delete(c.dataNodeStatuses, node)
	delete(c.blockExplorerStatuses, node)
	delete(c.nodeDownStatuses, node)
}

func (c *VegaMonitoringCollector) UpdateMonitoringDBStatuses(newStatuses read.MetaMonitoringStatuses) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.monitoringDatabaseStatuses = newStatuses
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

	// BlockExplorer
	ch <- desc.BlockExplorer.blockExplorerInfo

	// Node Down
	ch <- desc.NodeDown.nodeDown

	// MetaMonitoring: Monitoring Database
	ch <- desc.MonitoringDatabase.dataNodeData
}

// Collect returns the current state of all metrics of the collector.
func (c *VegaMonitoringCollector) Collect(ch chan<- prometheus.Metric) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.collectCoreStatuses(ch)
	c.collectDataNodeStatuses(ch)
	c.collectBlockExplorerStatuses(ch)
	c.collectNodeDownStatuses(ch)
	c.collectMonitoringDatabaseStatuses(ch)
}

func (c *VegaMonitoringCollector) collectCoreStatuses(ch chan<- prometheus.Metric) {
	for nodeName, nodeStatus := range c.coreStatuses {
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
	}
}

func (c *VegaMonitoringCollector) collectDataNodeStatuses(ch chan<- prometheus.Metric) {
	for nodeName, nodeStatus := range c.dataNodeStatuses {
		fieldToValue := map[*prometheus.Desc]float64{
			desc.Core.coreBlockHeight:         float64(nodeStatus.CoreBlockHeight),
			desc.DataNode.dataNodeBlockHeight: float64(nodeStatus.DataNodeBlockHeight),
			desc.Core.coreTime:                float64(nodeStatus.CoreTime.Unix()),
			desc.DataNode.dataNodeTime:        float64(nodeStatus.DataNodeTime.Unix()),
		}
		if nodeStatus.RESTReqDuration.Seconds() > 0 {
			fieldToValue[desc.DataNode.dataNodePerformanceRESTInfoDuration] = nodeStatus.RESTReqDuration.Seconds()
		}
		if nodeStatus.GQLReqDuration.Seconds() > 0 {
			fieldToValue[desc.DataNode.dataNodePerformanceGQLInfoDuration] = nodeStatus.GQLReqDuration.Seconds()
		}
		if nodeStatus.GRPCReqDuration.Seconds() > 0 {
			fieldToValue[desc.DataNode.dataNodePerformanceGRPCInfoDuration] = nodeStatus.GRPCReqDuration.Seconds()
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
}

func (c *VegaMonitoringCollector) collectNodeDownStatuses(ch chan<- prometheus.Metric) {

	for nodeName, nodeStatus := range c.nodeDownStatuses {
		// Block Explorer Down
		ch <- prometheus.MustNewConstMetric(
			desc.NodeDown.nodeDown, prometheus.UntypedValue, 1,
			// Labels
			nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
		)
	}
}

func (c *VegaMonitoringCollector) collectMonitoringDatabaseStatuses(ch chan<- prometheus.Metric) {

	twoMinutesAgo := time.Now().Add(-2 * time.Minute)

	if twoMinutesAgo.Before(c.monitoringDatabaseStatuses.UpdateTime) {
		fieldToValue := map[*prometheus.Desc]*int32{
			desc.MonitoringDatabase.dataNodeData:               c.monitoringDatabaseStatuses.DataNodeData,
			desc.MonitoringDatabase.assetPricesData:            c.monitoringDatabaseStatuses.AssetPricesData,
			desc.MonitoringDatabase.blockSignersData:           c.monitoringDatabaseStatuses.BlockSignersData,
			desc.MonitoringDatabase.cometTxsData:               c.monitoringDatabaseStatuses.CometTxsData,
			desc.MonitoringDatabase.networkBalancesData:        c.monitoringDatabaseStatuses.NetworkBalancesData,
			desc.MonitoringDatabase.networkHistorySegmentsData: c.monitoringDatabaseStatuses.NetworkHistorySegmentsData,
		}

		for field, value := range fieldToValue {
			if value != nil {
				ch <- prometheus.NewMetricWithTimestamp(
					c.monitoringDatabaseStatuses.UpdateTime,
					prometheus.MustNewConstMetric(
						field, prometheus.GaugeValue, float64(*value),
						// no extra Labels
					))
			}
		}
	}
}
