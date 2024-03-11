package entities

import (
	"time"
)

type MonitoringServiceType string

const (
	BlockSignersSvc       MonitoringServiceType = "BLOCK_SIGNERS"
	SegmentsSvc           MonitoringServiceType = "SEGMENTS"
	CometTxsSvc           MonitoringServiceType = "COMET_TXS"
	NetworkBalancesSvc    MonitoringServiceType = "NETWORK_BALANCES"
	AssetPricesSvc        MonitoringServiceType = "ASSET_PRICES"
	PromEthereumCallsSvc  MonitoringServiceType = "PROMETHEUS_ETHEREUM_CALLS_SERVICE"
	PromEthNodeScannerSvc MonitoringServiceType = "PROMETHEUS_ETH_NODE_SCANNER"
	PromNodeScannerSvc    MonitoringServiceType = "PROMETHEUS_NODE_SCANNER"
	PromMetamonitoringSvc MonitoringServiceType = "PROMETHEUS_METAMONITORING"
)

var AllMonitoringServices = []MonitoringServiceType{
	BlockSignersSvc,
	SegmentsSvc,
	CometTxsSvc,
	NetworkBalancesSvc,
	AssetPricesSvc,
}

type MetaMonitoringStatusCheck struct {
	CheckName  string
	IsHealthy  int32
	LastUpdate *time.Time
}

type UnhealthyReason int

const (
	ReasonUnknown                  UnhealthyReason = 0
	ReasonMissingStatusFromService UnhealthyReason = 1
	ReasonNetworkIsNotUpToDate     UnhealthyReason = 2
	ReasonTargetConnectionFailure  UnhealthyReason = 3

	ReasonEthereumGetBalancesFailure          UnhealthyReason = 4
	ReasonEthereumContractCallFailure         UnhealthyReason = 5
	ReasonEthereumContractInvalidResponseType UnhealthyReason = 6
	ReasonEthereumContractEventFilterFailure  UnhealthyReason = 7
)

type MonitoringStatus struct {
	StatusTime      time.Time             `db:"status_time"`
	IsHealthy       bool                  `db:"is_healthy"`
	Service         MonitoringServiceType `db:"monitoring_service"`
	UnhealthyReason UnhealthyReason       `db:"unhealthy_reason"`
}

func (s MonitoringStatus) UnhealthyReasonString() string {
	return UnHealthyReasonString(s.UnhealthyReason)
}

func UnHealthyReasonString(reason UnhealthyReason) string {
	switch reason {
	case ReasonMissingStatusFromService:
		return "Missing status from the service"
	case ReasonNetworkIsNotUpToDate:
		return "Network is not up to date"
	case ReasonTargetConnectionFailure:
		return "Target connection failure"
	case ReasonEthereumGetBalancesFailure:
		return "Failed to get account balances from ethereum"
	case ReasonEthereumContractCallFailure:
		return "Failed to call ethereum smart contract"
	case ReasonEthereumContractEventFilterFailure:
		return "Failed to filter ethereum smart contract event"
	}

	return "Unknown reason"
}
