package collectors

import "github.com/prometheus/client_golang/prometheus"

var desc = struct {
	Core struct {
		coreBlockHeight *prometheus.Desc
		coreTime        *prometheus.Desc
		coreInfo        *prometheus.Desc
	}

	DataNode struct {
		dataNodeBlockHeight *prometheus.Desc
		dataNodeTime        *prometheus.Desc

		dataNodePerformanceRESTInfoDuration *prometheus.Desc
		dataNodePerformanceGQLInfoDuration  *prometheus.Desc
		dataNodePerformanceGRPCInfoDuration *prometheus.Desc
	}

	BlockExplorer struct {
		blockExplorerInfo *prometheus.Desc
	}

	NodeDown struct {
		nodeDown *prometheus.Desc
	}

	MonitoringDatabase struct {
		dataNodeData               *prometheus.Desc
		assetPricesData            *prometheus.Desc
		blockSignersData           *prometheus.Desc
		cometTxsData               *prometheus.Desc
		networkBalancesData        *prometheus.Desc
		networkHistorySegmentsData *prometheus.Desc
	}
}{}

func init() {
	//
	// Core
	//
	desc.Core.coreBlockHeight = prometheus.NewDesc(
		"core_block_height", "Current Block Height of Core", []string{"node", "type", "environment", "internal"}, nil,
	)
	desc.Core.coreTime = prometheus.NewDesc(
		"core_time", "Current Block Time of Core", []string{"node", "type", "environment", "internal"}, nil,
	)
	desc.Core.coreInfo = prometheus.NewDesc(
		"core_info", "Basic information about node", []string{"node", "type", "environment", "internal", "chain_id", "app_version", "app_version_hash"}, nil,
	)

	//
	// Data Node
	//
	desc.DataNode.dataNodeBlockHeight = prometheus.NewDesc(
		"datanode_block_height", "Current Block Height of Data-Node", []string{"node", "type", "environment", "internal"}, nil,
	)
	desc.DataNode.dataNodeTime = prometheus.NewDesc(
		"datanode_time", "Current Block Time of Data-Node", []string{"node", "type", "environment", "internal"}, nil,
	)

	desc.DataNode.dataNodePerformanceRESTInfoDuration = prometheus.NewDesc(
		"datanode_performance_rest_info_duration", "Duration of REST request to get info about node", []string{"node", "type", "environment", "internal"}, nil,
	)
	desc.DataNode.dataNodePerformanceGQLInfoDuration = prometheus.NewDesc(
		"datanode_performance_gql_info_duration", "Duration of GraphQL request to get info about node", []string{"node", "type", "environment", "internal"}, nil,
	)
	desc.DataNode.dataNodePerformanceGRPCInfoDuration = prometheus.NewDesc(
		"datanode_performance_grpc_info_duration", "Duration of gRPC request to get info about node", []string{"node", "type", "environment", "internal"}, nil,
	)

	//
	// Block Explorer
	//
	desc.BlockExplorer.blockExplorerInfo = prometheus.NewDesc(
		"blockexplorer_info", "Basic information about block explorer", []string{"node", "type", "environment", "internal", "version", "version_hash"}, nil,
	)

	//
	// Node Down
	//
	desc.NodeDown.nodeDown = prometheus.NewDesc(
		"node_down", "Node is not responsive", []string{"node", "type", "environment", "internal"}, nil,
	)

	//
	// Meta-Monitoring: Monitoring Database
	//
	desc.MonitoringDatabase.dataNodeData = prometheus.NewDesc(
		"monitoring_db_datanode_data", "DataNode data in Monitoring database is healthy", nil, nil,
	)
	desc.MonitoringDatabase.assetPricesData = prometheus.NewDesc(
		"monitoring_db_assetprices_data", "Asset Prices data in Monitoring database is healthy", nil, nil,
	)
	desc.MonitoringDatabase.blockSignersData = prometheus.NewDesc(
		"monitoring_db_blocksigners_data", "Block Signers data in Monitoring database is healthy", nil, nil,
	)
	desc.MonitoringDatabase.cometTxsData = prometheus.NewDesc(
		"monitoring_db_comettxs_data", "Comet Txs data in Monitoring database is healthy", nil, nil,
	)
	desc.MonitoringDatabase.networkBalancesData = prometheus.NewDesc(
		"monitoring_db_networkbalances_data", "Network Balances data in Monitoring database is healthy", nil, nil,
	)
	desc.MonitoringDatabase.networkHistorySegmentsData = prometheus.NewDesc(
		"monitoring_db_networkhistorysegments_data", "Network History Segments data in Monitoring database is healthy", nil, nil,
	)
}
