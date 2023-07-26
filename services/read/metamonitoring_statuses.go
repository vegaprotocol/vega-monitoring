package read

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type MetaMonitoringStatuses struct {
	DataNodeData               int32
	AssetPricesData            int32
	BlockSignersData           int32
	CometTxsData               int32
	NetworkBalancesData        int32
	NetworkHistorySegmentsData int32

	UpdateTime time.Time
}

func (s *ReadService) GetMetaMonitoringStatuses(ctx context.Context) (MetaMonitoringStatuses, error) {
	result := MetaMonitoringStatuses{}

	logger := s.log.With(zap.String("reader", "MetaMonitoringStatuses"))

	logger.Info("Read Meta-Monitoring Statuses from Monitoring Database")

	metamonitoringStatusesStore := s.storeReadService.NewMetamonitoringStatus()

	checks, err := metamonitoringStatusesStore.GetAll(ctx)
	if err != nil {
		return result, err
	}

	for _, check := range checks {

		switch check.CheckName {
		case "data_node":
			result.DataNodeData = check.IsHealthy
		case "asset_prices":
			result.AssetPricesData = check.IsHealthy
		case "block_signers":
			result.BlockSignersData = check.IsHealthy
		case "comet_txs":
			result.CometTxsData = check.IsHealthy
		case "network_balances":
			result.NetworkBalancesData = check.IsHealthy
		case "network_history_segments":
			result.NetworkHistorySegmentsData = check.IsHealthy
		default:
			logger.Error("Unknown check name", zap.String("check_name", check.CheckName))
		}
	}
	if len(checks) != 6 {
		logger.Error("Wrong number of checks", zap.Int("expected", 6), zap.Int("actual", len(checks)))
	}

	return result, nil
}
