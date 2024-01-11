package entities

import (
	"time"
)

type MonitoringServiceType string

const (
	BlockSignersSvc    MonitoringServiceType = "BLOCK_SIGNERS"
	SegmentsSvc        MonitoringServiceType = "SEGMENTS"
	CometTxsSvc        MonitoringServiceType = "COMET_TXS"
	NetworkBalancesSvc MonitoringServiceType = "NETWORK_BALANCES"
	AssetPricesSvc     MonitoringServiceType = "ASSET_PRICES"
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
)

type MonitoringStatus struct {
	StatusTime      time.Time
	IsHealthy       bool
	Service         MonitoringServiceType
	UnhealthyReason UnhealthyReason
}

func (s MonitoringStatus) UnhealthyReasonString() string {
	return UnHealthyReasonString(s.UnhealthyReason)
}

func UnHealthyReasonString(reason UnhealthyReason) string {

	return "Unknown reason"
}
