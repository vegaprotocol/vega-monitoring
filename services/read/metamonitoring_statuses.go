package read

import (
	"context"
	"time"

	"github.com/vegaprotocol/vega-monitoring/entities"
	"go.uber.org/zap"
)

type MetaMonitoringStatuses struct {
	DataNodeData               *int32
	AssetPricesData            *int32
	BlockSignersData           *int32
	CometTxsData               *int32
	NetworkBalancesData        *int32
	NetworkHistorySegmentsData *int32

	UpdateTime time.Time
}

func (s *ReadService) GetMetaMonitoringStatuses(ctx context.Context) (MetaMonitoringStatuses, error) {
	var permanentOne int32 = 1
	result := MetaMonitoringStatuses{
		DataNodeData: &permanentOne,
	}

	logger := s.log.With(zap.String("reader", "MetaMonitoringStatuses"))

	logger.Info("Read Meta-Monitoring Statuses from Monitoring Database")

	metamonitoringStatusesStore := s.storeReadService.NewMonitoringStatus()

	checks, err := metamonitoringStatusesStore.GetLatest(ctx)
	if err != nil {
		return result, err
	}

	result.UpdateTime = time.Now()

	for _, check := range checks {
		var isHealthyMetricsValue int32 = 0
		if check.IsHealthy {
			isHealthyMetricsValue = 1
		}
		switch check.Service {
		// case "data_node":
		// 	result.DataNodeData = &isHealthyMetricsValue
		case entities.AssetPricesSvc:
			result.AssetPricesData = &isHealthyMetricsValue
		case entities.BlockSignersSvc:
			result.BlockSignersData = &isHealthyMetricsValue
		case entities.CometTxsSvc:
			result.CometTxsData = &isHealthyMetricsValue
		case entities.NetworkBalancesSvc:
			result.NetworkBalancesData = &isHealthyMetricsValue
		case entities.SegmentsSvc:
			result.NetworkHistorySegmentsData = &isHealthyMetricsValue
		default:
			logger.Error("Unknown check name", zap.String("check_name", string(check.Service)))
		}
	}

	if len(checks) != 6 {
		logger.Error("Wrong number of checks", zap.Int("expected", 6), zap.Int("actual", len(checks)), zap.Any("checks", checks))
	}

	return result, nil
}
