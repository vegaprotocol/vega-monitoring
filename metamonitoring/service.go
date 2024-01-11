package metamonitoring

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
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

type MonitoringStatusUpdateService struct {
	store          *sqlstore.MonitoringStatus
	logger         *logging.Logger
	activeServices []entities.MonitoringServiceType
}

func NewMonitoringStatusUpdateService(store *sqlstore.MonitoringStatus, logger *logging.Logger) (*MonitoringStatusUpdateService, error) {
	return &MonitoringStatusUpdateService{
		store:  store,
		logger: logger,
	}, nil
}

func (msus *MonitoringStatusUpdateService) BlockSignersStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.BlockSignersSvc)

	return &monitoringStatusPublisherService{
		store:   msus.store,
		service: entities.BlockSignersSvc,
	}
}

func (msus *MonitoringStatusUpdateService) SegmentsStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.SegmentsSvc)

	return &monitoringStatusPublisherService{
		store:   msus.store,
		service: entities.SegmentsSvc,
	}
}

func (msus *MonitoringStatusUpdateService) CometTxsStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.CometTxsSvc)

	return &monitoringStatusPublisherService{
		store:   msus.store,
		service: entities.CometTxsSvc,
	}
}

func (msus *MonitoringStatusUpdateService) NetworkBalancesStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.NetworkBalancesSvc)

	return &monitoringStatusPublisherService{
		store:   msus.store,
		service: entities.NetworkBalancesSvc,
	}
}

func (msus *MonitoringStatusUpdateService) AssetPricesStatusPublisher() MonitoringStatusPublisher {
	msus.activeServices = append(msus.activeServices, entities.AssetPricesSvc)

	return &monitoringStatusPublisherService{
		store:   msus.store,
		service: entities.AssetPricesSvc,
	}
}

func (msus *MonitoringStatusUpdateService) Run(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			// Check if all of the monitoring services provided any status update
			// if not add failed state
			for _, service := range msus.activeServices {
				if !msus.store.IsPendingFor(service) {
					msus.store.Add(entities.MonitoringStatus{
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

			ctx, cancel := context.WithCancel(ctx)
			// Upsert all the pending states
			if _, err := msus.store.FlushUpsert(ctx); err != nil {
				msus.logger.Errorf("failed to flush upsert monitoring status updates: %w", err)
			}
			cancel()
		case <-ctx.Done():
			return
		}
	}
}
