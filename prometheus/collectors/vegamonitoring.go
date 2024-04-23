package collectors

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
	"github.com/vegaprotocol/vega-monitoring/services/read"
)

type AccountBalanceMetric struct {
	Value     float64
	ChainId   string
	NetworkId string
	NodeName  string
}

type ContractCallResponse struct {
	Value           float64
	ContractAddress string
	MethodName      string
	NodeName        string
}

type VegaMonitoringCollector struct {
	coreStatuses            map[string]*types.CoreStatus
	dataNodeStatuses        map[string]*types.DataNodeStatus
	blockExplorerStatuses   map[string]*types.BlockExplorerStatus
	ethereumAccountBalances map[string]AccountBalanceMetric
	contractCallResponse    map[string]ContractCallResponse

	// Meta-Monitoring
	monitoringDatabaseStatuses read.MetaMonitoringStatuses

	// Ethereum Node Statuses
	ethNodeStatuses []types.EthereumNodeStatus
	ethNodeHeights  map[string]types.EthereumNodeHeight
	contractEvents  map[types.EntityHash]types.EthereumContractsEvents

	accessMu sync.RWMutex
}

func NewVegaMonitoringCollector() *VegaMonitoringCollector {
	return &VegaMonitoringCollector{
		coreStatuses:            map[string]*types.CoreStatus{},
		dataNodeStatuses:        map[string]*types.DataNodeStatus{},
		blockExplorerStatuses:   map[string]*types.BlockExplorerStatus{},
		ethereumAccountBalances: map[string]AccountBalanceMetric{},
		contractCallResponse:    map[string]ContractCallResponse{},
		ethNodeHeights:          map[string]types.EthereumNodeHeight{},

		contractEvents: map[types.EntityHash]types.EthereumContractsEvents{},
	}
}

func (c *VegaMonitoringCollector) UpdateCoreStatus(node string, newStatus *types.CoreStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.coreStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) UpdateEthereumAccountBalance(
	nodeName string,
	accountAddress string,
	chainId, networkId string,
	val float64,
) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()

	c.ethereumAccountBalances[accountAddress] = AccountBalanceMetric{
		NodeName:  nodeName,
		NetworkId: networkId,
		ChainId:   chainId,
		Value:     val,
	}
}

func (c *VegaMonitoringCollector) UpdateEthereumCallResponse(
	nodeName string,
	id string,
	contractAddress string,
	methodName string,
	val float64,
) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()

	c.contractCallResponse[id] = ContractCallResponse{
		Value:           val,
		ContractAddress: contractAddress,
		MethodName:      methodName,
		NodeName:        nodeName,
	}
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

func (c *VegaMonitoringCollector) clearStatusFor(node string) {
	delete(c.coreStatuses, node)
	delete(c.dataNodeStatuses, node)
	delete(c.blockExplorerStatuses, node)
}

func (c *VegaMonitoringCollector) UpdateMonitoringDBStatuses(newStatuses read.MetaMonitoringStatuses) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.monitoringDatabaseStatuses = newStatuses
}

func (c *VegaMonitoringCollector) UpdateEthereumNodeStatuses(nodeHealthy []types.EthereumNodeStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.ethNodeStatuses = nodeHealthy
}

func (c *VegaMonitoringCollector) UpdateEthereumNodeHeights(heights []types.EthereumNodeHeight) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()

	for idx, item := range heights {
		c.ethNodeHeights[item.RPCEndpoint] = heights[idx]
	}
}

func (c *VegaMonitoringCollector) UpdateEthereumContractEvents(events []types.EthereumContractsEvents) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()

	for idx, event := range events {
		// Make sure events are not duplicated in the /metrics page
		c.contractEvents[event.Hash()] = events[idx]
	}
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
	ch <- desc.DataNode.dataNodeScore
	ch <- desc.DataNode.dataNodePerformanceRESTInfoDuration
	ch <- desc.DataNode.dataNodePerformanceGQLInfoDuration
	ch <- desc.DataNode.dataNodePerformanceGRPCInfoDuration

	// BlockExplorer
	ch <- desc.BlockExplorer.blockExplorerInfo

	// MetaMonitoring: Monitoring Database
	ch <- desc.MetaMonitoring.monitoringDatabaseHealthy

	// Ethereum Node Statuses
	ch <- desc.EthereumNodeStatus
	ch <- desc.EthereumNodeHeight

	// Ethereum on chain data
	ch <- desc.EthereumAccountBalances
	ch <- desc.EthereumContractCallResponse
	ch <- desc.EthereumContractEvents
}

// Collect returns the current state of all metrics of the collector.
func (c *VegaMonitoringCollector) Collect(ch chan<- prometheus.Metric) {
	// TODO(fixme): Is it good idea to lock access mutex here, when We do not know whats going on in child functions?
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.collectCoreStatuses(ch)
	c.collectDataNodeStatuses(ch)
	c.collectBlockExplorerStatuses(ch)
	c.collectMonitoringDatabaseStatuses(ch)
	c.collectEthereumNodeStatuses(ch)
	c.collectEthereumNodesHeights(ch)
	c.collectEthereumAccountBalances(ch)
	c.collectEthereumContractCallResponses(ch)
	c.collectEthereumContractEvents(ch)
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
		// Data Node Score
		ch <- prometheus.NewMetricWithTimestamp(
			nodeStatus.CurrentTime,
			prometheus.MustNewConstMetric(
				desc.DataNode.dataNodeScore, prometheus.GaugeValue, float64(nodeStatus.GetScore()),
				// Labels
				nodeName, string(nodeStatus.Type), nodeStatus.Environment, strconv.FormatBool(nodeStatus.Internal),
				// Extra labels
				strconv.FormatUint(nodeStatus.GRPCScore, 10), strconv.FormatUint(nodeStatus.RESTScore, 10),
				strconv.FormatUint(nodeStatus.GQLScore, 10), strconv.FormatUint(nodeStatus.GetUpToDateScore(), 10),
				strconv.FormatUint(nodeStatus.Data1DayScore, 10), strconv.FormatUint(nodeStatus.Data1WeekScore, 10),
				strconv.FormatUint(nodeStatus.DataArchivalScore, 10),
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

func (c *VegaMonitoringCollector) collectMonitoringDatabaseStatuses(ch chan<- prometheus.Metric) {

	twoMinutesAgo := time.Now().Add(-2 * time.Minute)

	if twoMinutesAgo.Before(c.monitoringDatabaseStatuses.UpdateTime) {
		fieldToValue := map[string]*int32{
			"data_node":                c.monitoringDatabaseStatuses.DataNodeData,
			"asset_prices":             c.monitoringDatabaseStatuses.AssetPricesData,
			"block_signers":            c.monitoringDatabaseStatuses.BlockSignersData,
			"comet_txs":                c.monitoringDatabaseStatuses.CometTxsData,
			"network_balances":         c.monitoringDatabaseStatuses.NetworkBalancesData,
			"network_history_segments": c.monitoringDatabaseStatuses.NetworkHistorySegmentsData,
		}

		for dataType, value := range fieldToValue {
			if value != nil {
				ch <- prometheus.NewMetricWithTimestamp(
					c.monitoringDatabaseStatuses.UpdateTime,
					prometheus.MustNewConstMetric(
						desc.MetaMonitoring.monitoringDatabaseHealthy, prometheus.GaugeValue, float64(*value),
						// Labels
						dataType,
					))
			}
		}
	}
}

func (c *VegaMonitoringCollector) collectEthereumNodeStatuses(ch chan<- prometheus.Metric) {
	for _, ethNodeStatus := range c.ethNodeStatuses {
		status := 1.0
		if !ethNodeStatus.Healthy {
			status = 0
		}

		ch <- prometheus.NewMetricWithTimestamp(
			ethNodeStatus.UpdateTime,
			prometheus.MustNewConstMetric(
				desc.EthereumNodeStatus, prometheus.GaugeValue, status,
				// Labels
				ethNodeStatus.NodeName, ethNodeStatus.ChainId, ethNodeStatus.RPCEndpoint,
			))
	}
}

func (c *VegaMonitoringCollector) collectEthereumNodesHeights(ch chan<- prometheus.Metric) {
	for _, metric := range c.ethNodeHeights {
		ch <- prometheus.NewMetricWithTimestamp(
			metric.UpdateTime,
			prometheus.MustNewConstMetric(
				desc.EthereumNodeHeight, prometheus.CounterValue, float64(metric.Height),
				// Labels
				metric.NodeName, metric.ChainId, metric.RPCEndpoint,
			),
		)
	}
}

func (c *VegaMonitoringCollector) collectEthereumAccountBalances(ch chan<- prometheus.Metric) {
	for accAddress, metric := range c.ethereumAccountBalances {
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				desc.EthereumAccountBalances, prometheus.GaugeValue, metric.Value,
				// Labels
				metric.NodeName, metric.NetworkId, metric.ChainId, accAddress,
			),
		)
	}
}

func (c *VegaMonitoringCollector) collectEthereumContractCallResponses(ch chan<- prometheus.Metric) {
	for id, metric := range c.contractCallResponse {
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				desc.EthereumContractCallResponse, prometheus.GaugeValue, metric.Value,
				// Labels
				metric.NodeName, id, metric.ContractAddress, metric.MethodName,
			),
		)
	}
}

func (c *VegaMonitoringCollector) collectEthereumContractEvents(ch chan<- prometheus.Metric) {
	for _, metric := range c.contractEvents {
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				desc.EthereumContractEvents, prometheus.CounterValue, float64(metric.Count),
				// Labels
				metric.NodeName, metric.ID, metric.ContractAddress, metric.EventName,
			),
		)
	}
}
