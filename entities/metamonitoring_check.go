package entities

import (
	"time"
)

type MetaMonitoringStatusCheck struct {
	CheckName  string
	IsHealthy  int32
	LastUpdate time.Time
}
