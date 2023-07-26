package metamonitoring

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"go.uber.org/zap"
)

type MetaMonitoringStatusService struct {
	store     *read.ReadService
	collector *collectors.VegaMonitoringCollector
	log       *logging.Logger
}

func NewMetaMonitoringStatusService(
	store *read.ReadService,
	collector *collectors.VegaMonitoringCollector,
	log *logging.Logger,
) *MetaMonitoringStatusService {
	log = log.With(zap.String("service", "metamonitoring-status"))

	return &MetaMonitoringStatusService{
		store:     store,
		collector: collector,
		log:       log,
	}
}

func (s *MetaMonitoringStatusService) Start(ctx context.Context) error {

	s.startUpdatingMetamonitoringStatuses(ctx)

	return nil
}

func (s *MetaMonitoringStatusService) startUpdatingMetamonitoringStatuses(ctx context.Context) {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		statuses, err := s.store.GetMetaMonitoringStatuses(ctx)

		if err != nil {
			s.log.Error("Failed to get Meta-Monitoring statuses from monitoring database", zap.Error(err))
			continue
		}

		s.collector.UpdateMonitoringDBStatuses(statuses)

		s.log.Debug("successfully updated Meta-Monitoring statuses in prometheus")

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Scanning Cores for Prometheus")
			return
		case <-ticker.C:
			continue
		}
	}
}
