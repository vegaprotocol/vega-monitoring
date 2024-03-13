package metamonitoring

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	vegaclient "github.com/vegaprotocol/vega-monitoring/clients/vega"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

const (
	BlockDiffHealthyThreshold = 900
	TimeDiffHealthyThreshold  = 10 * time.Minute
)

type MonitoringStatusPublisher interface {
	Publish(isHealthy bool) error
	PublishWithReason(isHealthy bool, reason entities.UnhealthyReason) error
}

type monitoringStatusPublisherService struct {
	store *sqlstore.MonitoringStatus

	service entities.MonitoringServiceType
}

func (ps *monitoringStatusPublisherService) Publish(isHealthy bool) error {
	event := entities.MonitoringStatus{
		StatusTime:      time.Now(),
		IsHealthy:       isHealthy,
		Service:         ps.service,
		UnhealthyReason: entities.ReasonUnknown,
	}

	ps.store.Add(event)

	return nil
}

func (ps *monitoringStatusPublisherService) PublishWithReason(isHealthy bool, reason entities.UnhealthyReason) error {
	event := entities.MonitoringStatus{
		StatusTime:      time.Now(),
		IsHealthy:       isHealthy,
		Service:         ps.service,
		UnhealthyReason: reason,
	}

	ps.store.Add(event)

	return nil
}

type MonitoringStore interface {
	NewMonitoringStatus() *sqlstore.MonitoringStatus
	NewBlocks() *sqlstore.Blocks
}

type VegaClient interface {
	GetStatistics() (*vegaclient.Statistics, error)
}

type MetamonitoringService interface {
	BlockSignersStatusPublisher() MonitoringStatusPublisher
	SegmentsStatusPublisher() MonitoringStatusPublisher
	CometTxsStatusPublisher() MonitoringStatusPublisher
	NetworkBalancesStatusPublisher() MonitoringStatusPublisher
	AssetPricesStatusPublisher() MonitoringStatusPublisher
	PrometheusEthereumCalls() MonitoringStatusPublisher
	Run(ctx context.Context, tickInterval time.Duration)
}

type MonitoringStatusUpdateService struct {
	monitoringStatusStore *sqlstore.MonitoringStatus
	blocksStore           *sqlstore.Blocks
	vegaClient            VegaClient
	logger                *logging.Logger
	activeServices        []entities.MonitoringServiceType
}
