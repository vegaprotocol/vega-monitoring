package entities

import (
	"time"
)

type MetaMonitoringStatusCheck struct {
	CheckName  string
	IsHealthy  int32
	LastUpdate *time.Time
}

type UnhealthyReason int

type MonitoringStatus struct {
	StatusTime      time.Time
	IsHealthy       bool
	Service         ServiceType
	UnhealthyReason UnhealthyReason
}

func (s MonitoringStatus) UnhealthyReasonString() string {
	return UnHealthyReasonString(s.UnhealthyReason)
}

func UnHealthyReasonString(reason UnhealthyReason) string {

	return "Unknown reason"
}
