package read

import (
	"context"
	"time"

	"github.com/vegaprotocol/vega-monitoring/entities"
	"go.uber.org/zap"
)

const expectedChecks = 7

type MetaMonitoringStatuses struct {
	AssetPricesData             *int32
	BlockSignersData            *int32
	CometTxsData                *int32
	NetworkBalancesData         *int32
	NetworkHistorySegmentsData  *int32
	PrometheusEthereumCallsData *int32

	// Unused
	DataNodeData             *int32
	PrometheusEthNodeScanner *int32
	PrometheusNodeScanner    *int32
	PrometheusMetamonitoring *int32

	UpdateTime time.Time
}

type StatusDetails struct {
	Healthy         bool
	UpdatedAt       time.Time
	UnhealthyReason entities.UnhealthyReason
}

type MetaMonitoringStatusesExtended struct {
	HealthyOverAll              bool
	AssetPricesData             StatusDetails
	BlockSignersData            StatusDetails
	CometTxsData                StatusDetails
	NetworkBalancesData         StatusDetails
	NetworkHistorySegmentsData  StatusDetails
	PrometheusEthereumCallsData StatusDetails

	// Unused
	DataNodeData                 StatusDetails
	PrometheusEthNodeScannerData StatusDetails
	PrometheusNodeScannerData    StatusDetails
	PrometheusMetamonitoringData StatusDetails
}

func EmptyMetaMonitoringStatusesExtended() *MetaMonitoringStatusesExtended {
	return &MetaMonitoringStatusesExtended{
		HealthyOverAll: false,
		DataNodeData: StatusDetails{
			Healthy:         true, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		AssetPricesData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		BlockSignersData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		CometTxsData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		NetworkBalancesData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		NetworkHistorySegmentsData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		PrometheusEthereumCallsData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		PrometheusEthNodeScannerData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		PrometheusNodeScannerData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
		PrometheusMetamonitoringData: StatusDetails{
			Healthy:         false, // We do not use this check anymore
			UpdatedAt:       time.Unix(0, 0),
			UnhealthyReason: entities.ReasonUnknown,
		},
	}
}

func (s *ReadService) GetMetaMonitoringStatuses(ctx context.Context) (MetaMonitoringStatuses, error) {
	var permanentOne int32 = 1
	result := MetaMonitoringStatuses{
		PrometheusEthNodeScanner: &permanentOne,
		PrometheusNodeScanner:    &permanentOne,
		PrometheusMetamonitoring: &permanentOne,
	}

	logger := s.log.With(zap.String("reader", "MetaMonitoringStatuses"))

	logger.Debug("Read Meta-Monitoring Statuses from Monitoring Database")

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
		case entities.DataNodeSvc:
			result.DataNodeData = &isHealthyMetricsValue
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
		case entities.PromEthereumCallsSvc:
			result.PrometheusEthereumCallsData = &isHealthyMetricsValue
		default:
			logger.Error("Unknown check name", zap.String("check_name", string(check.Service)))
		}
	}

	if len(checks) != expectedChecks {
		logger.Error("Wrong number of checks", zap.Int("expected", expectedChecks), zap.Int("actual", len(checks)), zap.Any("checks", checks))
	}

	return result, nil
}

func (s *ReadService) GetMetaMonitoringStatusesExtended(ctx context.Context) (*MetaMonitoringStatusesExtended, error) {
	result := EmptyMetaMonitoringStatusesExtended()

	logger := s.log.With(zap.String("reader", "GetMetaMonitoringStatusesExtended"))

	logger.Debug("Read Meta-Monitoring-Extended Statuses from Monitoring Database")

	metamonitoringStatusesStore := s.storeReadService.NewMonitoringStatus()

	checks, err := metamonitoringStatusesStore.GetLatest(ctx)
	if err != nil {
		return result, err
	}

	for _, check := range checks {
		switch check.Service {
		// case "data_node":
		// 	result.DataNodeData = &isHealthyMetricsValue
		case entities.AssetPricesSvc:
			result.AssetPricesData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}
		case entities.BlockSignersSvc:
			result.BlockSignersData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}
		case entities.CometTxsSvc:
			result.CometTxsData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}
		case entities.NetworkBalancesSvc:
			result.NetworkBalancesData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}
		case entities.SegmentsSvc:
			result.NetworkHistorySegmentsData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}

		case entities.PromEthereumCallsSvc:
			result.PrometheusEthereumCallsData = StatusDetails{
				Healthy:         check.IsHealthy,
				UpdatedAt:       check.StatusTime,
				UnhealthyReason: check.UnhealthyReason,
			}

		// Unused
		default:
			logger.Error("Unknown check name", zap.String("check_name", string(check.Service)))
		}
	}

	result.HealthyOverAll = result.AssetPricesData.Healthy &&
		result.BlockSignersData.Healthy &&
		result.CometTxsData.Healthy &&
		result.NetworkBalancesData.Healthy &&
		result.NetworkHistorySegmentsData.Healthy &&
		result.PrometheusEthereumCallsData.Healthy

	if len(checks) != 5 {
		logger.Error("Wrong number of checks", zap.Int("expected", 6), zap.Int("actual", len(checks)), zap.Any("checks", checks))
	}

	return result, nil
}
