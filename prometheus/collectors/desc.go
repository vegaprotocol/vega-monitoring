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
		dataNodeScore       *prometheus.Desc

		dataNodePerformanceRESTInfoDuration *prometheus.Desc
		dataNodePerformanceGQLInfoDuration  *prometheus.Desc
		dataNodePerformanceGRPCInfoDuration *prometheus.Desc
	}

	BlockExplorer struct {
		blockExplorerInfo *prometheus.Desc
	}

	MetaMonitoring struct {
		monitoringDatabaseHealthy *prometheus.Desc
	}

	EthereumNodeStatus *prometheus.Desc
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
	desc.DataNode.dataNodeScore = prometheus.NewDesc(
		"datanode_score", "Cumulative score of Data-Node APIs: gRPC, REST, GraphQL", []string{"node", "type", "environment", "internal"}, nil,
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
	// Meta-Monitoring: Monitoring Database
	//
	desc.MetaMonitoring.monitoringDatabaseHealthy = prometheus.NewDesc(
		"monitoring_db_status", "Status of data in Monitoring Database. 1 good, 0 bad", []string{"data_type"}, nil,
	)

	//
	// Ethereum Nodes Statuses
	//
	desc.EthereumNodeStatus = prometheus.NewDesc(
		"ethereum_node_status", "Status of an Ethereum Node. 1 good, 0 bad", []string{"rpc_endpoint"}, nil,
	)
}
