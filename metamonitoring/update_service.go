package metamonitoring

import (
	"context"
	"fmt"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"go.uber.org/zap"
)

func NewMonitoringStatusUpdateService(store MonitoringStore, vegaClient VegaClient, logger *logging.Logger) (MetamonitoringService, error) {
	return &MonitoringStatusUpdateService{
		monitoringStatusStore: store.NewMonitoringStatus(),
		blocksStore:           store.NewBlocks(),
		vegaClient:            vegaClient,
		logger:                logger,
	}, nil
}

func (msus *MonitoringStatusUpdateService) DataNodeStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.DataNodeSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.DataNodeSvc,
	}
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

func (msus *MonitoringStatusUpdateService) PrometheusEthereumCalls() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.AssetPricesSvc)

	return &monitoringStatusPublisherService{
		store:   msus.monitoringStatusStore,
		service: entities.PromEthereumCallsSvc,
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

	latestDataNodeBlock, err := msus.blocksStore.GetLatestBlockWithCache(ctx, 30*time.Second)
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
						UnhealthyReason: entities.ReasonNodeIsNotUpToDate,
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
			msus.logger.Error("failed to flush upsert monitoring status updates", zap.Error(err))
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
