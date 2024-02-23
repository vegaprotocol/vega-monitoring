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
}

type ContractCallResponse struct {
	Value           float64
	ContractAddress string
	MethodName      string
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
	ethNodeStatuses types.EthereumNodeStatuses

	accessMu sync.RWMutex
}

func NewVegaMonitoringCollector() *VegaMonitoringCollector {
	return &VegaMonitoringCollector{
		coreStatuses:            map[string]*types.CoreStatus{},
		dataNodeStatuses:        map[string]*types.DataNodeStatus{},
		blockExplorerStatuses:   map[string]*types.BlockExplorerStatus{},
		ethereumAccountBalances: map[string]AccountBalanceMetric{},
		contractCallResponse:    map[string]ContractCallResponse{},
	}
}

func (c *VegaMonitoringCollector) UpdateCoreStatus(node string, newStatus *types.CoreStatus) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.clearStatusFor(node)
	c.coreStatuses[node] = newStatus
}

func (c *VegaMonitoringCollector) UpdateEthereumAccountBalance(accountAddress string, chainId, networkId string, val float64) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()

	c.ethereumAccountBalances[accountAddress] = AccountBalanceMetric{
		NetworkId: networkId,
		ChainId:   chainId,
		Value:     val,
	}
}

func (c *VegaMonitoringCollector) UpdateEthereumCallResponse(
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

func (c *VegaMonitoringCollector) UpdateEthereumNodeStatuses(nodeHealthy map[string]bool, updateTime time.Time) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.ethNodeStatuses = types.EthereumNodeStatuses{
		NodeHealthy: nodeHealthy,
		UpdateTime:  updateTime,
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

	// Ethereum on chain data
	ch <- desc.EthereumAccountBalances
	ch <- desc.EthereumContractCallResponse
}

// Collect returns the current state of all metrics of the collector.
func (c *VegaMonitoringCollector) Collect(ch chan<- prometheus.Metric) {
	c.accessMu.Lock()
	defer c.accessMu.Unlock()
	c.collectCoreStatuses(ch)
	c.collectDataNodeStatuses(ch)
	c.collectBlockExplorerStatuses(ch)
	c.collectMonitoringDatabaseStatuses(ch)
	c.collectEthereumNodeStatuses(ch)
	c.collectEthereumAccountBalances(ch)
	c.collectEthereumContractCallResponses(ch)
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

	for ethNodeName, ethNodeStatus := range c.ethNodeStatuses.NodeHealthy {
		status := 1.0
		if !ethNodeStatus {
			status = 0
		}
		ch <- prometheus.NewMetricWithTimestamp(
			c.ethNodeStatuses.UpdateTime,
			prometheus.MustNewConstMetric(
				desc.EthereumNodeStatus, prometheus.GaugeValue, status,
				// Labels
				ethNodeName,
			))
	}
}

func (c *VegaMonitoringCollector) collectEthereumAccountBalances(ch chan<- prometheus.Metric) {
	for accAddress, metric := range c.ethereumAccountBalances {
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				desc.EthereumAccountBalances, prometheus.GaugeValue, metric.Value,
				// Labels
				metric.NetworkId, metric.ChainId, accAddress,
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
				id, metric.ContractAddress, metric.MethodName,
			),
		)
	}
}
