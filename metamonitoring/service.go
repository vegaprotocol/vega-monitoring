package metamonitoring

import (
	"context"
	"fmt"
	"time"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	vegaclient "github.com/vegaprotocol/vega-monitoring/clients/vega"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
	"go.uber.org/zap"
)

const (
	BlockDiffHealthyThreshold = 900
	TimeDiffHealthyThreshold  = 10 * time.Minute
)

type MonitoringStatusPublisher interface {
	Publish(isHealthy bool) error
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

type MonitoringStore interface {
	NewMonitoringStatus() *sqlstore.MonitoringStatus
	NewBlocks() *vega_sqlstore.Blocks
}

type VegaClient interface {
	GetStatistics() (*vegaclient.Statistics, error)
}

type MonitoringStatusUpdateService struct {
	monitoringStatusStore *sqlstore.MonitoringStatus
	blocksStore           *vega_sqlstore.Blocks
	vegaClient            VegaClient
	logger                *logging.Logger
	activeServices        []entities.MonitoringServiceType
}

func NewMonitoringStatusUpdateService(store MonitoringStore, vegaClient VegaClient, logger *logging.Logger) (*MonitoringStatusUpdateService, error) {
	return &MonitoringStatusUpdateService{
		monitoringStatusStore: store.NewMonitoringStatus(),
		blocksStore:           store.NewBlocks(),
		vegaClient:            vegaClient,
		logger:                logger,
	}, nil
}

func (msus *MonitoringStatusUpdateService) BlockSignersStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.BlockSignersSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.BlockSignersSvc,
	}
}

func (msus *MonitoringStatusUpdateService) SegmentsStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.SegmentsSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.SegmentsSvc,
	}
}

func (msus *MonitoringStatusUpdateService) CometTxsStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.CometTxsSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.CometTxsSvc,
	}
}

func (msus *MonitoringStatusUpdateService) NetworkBalancesStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.NetworkBalancesSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.NetworkBalancesSvc,
	}
}

func (msus *MonitoringStatusUpdateService) AssetPricesStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.AssetPricesSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.AssetPricesSvc,
	}
}

func (msus *MonitoringStatusUpdateService) isNodeUpToDate(ctx context.Context) (bool, error) {
	coreStatistics, err := msus.vegaClient.GetStatistics()
	if err != nil {
		return false, fmt.Errorf("failed to get core statistics: %w", err)
	}

	timeDiff := coreStatistics.CurrentTime.Sub(coreStatistics.VegaTime)
	if timeDiff > TimeDiffHealthyThreshold {
		msus.logger.Warningf(
			"Local node is not up to date: (currentTime(%s) - vegaTime(%s)) > %s. Diff %s",
			coreStatistics.CurrentTime.String(),
			coreStatistics.VegaTime.String(),
			TimeDiffHealthyThreshold.String(),
			timeDiff.String(),
		)
		return false, nil
	}

	latestDataNodeBlock, err := msus.blocksStore.GetLastBlock(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get last block: %w", err)
	}

	blocksDiff := coreStatistics.BlockHeight - latestDataNodeBlock.Height
	if blocksDiff > BlockDiffHealthyThreshold {
		msus.logger.Warningf(
			"Local node is not up to date: (vega block(%d) - datanode block(%d)) > %d. Diff %d",
			coreStatistics.BlockHeight,
			latestDataNodeBlock.Height,
			BlockDiffHealthyThreshold,
			blocksDiff,
		)
		return false, nil
	}

	return true, nil
}

func (msus *MonitoringStatusUpdateService) Run(ctx context.Context, tickInterval time.Duration) {
	time.Sleep(tickInterval + 3)
	ticker := time.NewTicker(tickInterval)

	monitoringStatusStore := msus.monitoringStatusStore
	for {
		isUpToDate, err := msus.isNodeUpToDate(ctx)
		if err != nil {
			msus.logger.Errorf("failed to check if node is up to date: %s", err.Error())
		}

		// Node is lagging too much or it is still replaying
		// in this case we publish failed state with the
		// correct error
		if !isUpToDate {
			msus.logger.Warningf("The local vega node is not up to date. Failing all the checks.")
			_, err := monitoringStatusStore.FlushClear(ctx)
			if err != nil {
				msus.logger.Error("failed to flush clear all actual health checks when node is not up to date", zap.Error(err))
			}

			for _, service := range msus.activeServices {
				if !monitoringStatusStore.IsPendingFor(service) {
					monitoringStatusStore.Add(entities.MonitoringStatus{
						StatusTime:      time.Now(),
						IsHealthy:       false,
						Service:         service,
						UnhealthyReason: entities.ReasonNetworkIsNotUpToDate,
					})
				}
			}
		}

		// Check if all of the monitoring services provided any status update
		// if not add failed state
		for _, service := range msus.activeServices {
			if !monitoringStatusStore.IsPendingFor(service) {
				monitoringStatusStore.Add(entities.MonitoringStatus{
					StatusTime:      time.Now(),
					IsHealthy:       false,
					Service:         service,
					UnhealthyReason: entities.ReasonMissingStatusFromService,
				})

				msus.logger.Debugf(
					"Service %s did not published status update for last 5 min. Marking it as unhealthy with the %s reason",
					service,
					entities.ReasonMissingStatusFromService,
				)
			}
		}

		innerCtx, cancel := context.WithCancel(ctx)
		// Upsert all the pending states
		if _, err := monitoringStatusStore.FlushUpsert(innerCtx); err != nil {
			msus.logger.Errorf("failed to flush upsert monitoring status updates: %w", err)
		}

		cancel()
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}