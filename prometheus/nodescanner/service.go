package nodescanner

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"go.uber.org/zap"
)

type NodeScannerService struct {
	config    *config.MonitoringConfig
	collector *collectors.VegaMonitoringCollector
	log       *logging.Logger
}

func NewNodeScannerService(
	config *config.MonitoringConfig,
	collector *collectors.VegaMonitoringCollector,
	log *logging.Logger,
) *NodeScannerService {
	log = log.With(zap.String("service", "node-scanner"))

	return &NodeScannerService{
		config:    config,
		collector: collector,
		log:       log,
	}
}

func (s *NodeScannerService) Start(ctx context.Context) error {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		// Go DataNode one-by-one synchroniusly
		for _, node := range s.config.DataNode {
			s.log.Debug("getting data for data-node", zap.String("name", node.Name))
			dataNodeStats, err := requestDataNodeStats(node.REST)
			if err != nil {
				s.log.Error("Failed to get data for data-node", zap.String("node", node.Name), zap.Error(err))
				s.collector.UpdateDataNodeStatusAsError(node.Name, err)
			} else {
				dataNodeStats.RESTReqDuration, _ = checkREST(node.REST)
				dataNodeStats.GQLReqDuration, _ = checkGQL(node.GraphQL)
				dataNodeStats.GRPCReqDuration, _ = checkGRPC(node.GRPC)
				s.collector.UpdateDataNodeStatus(node.Name, dataNodeStats)
			}
			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		// Go BlockExplorer one-by-one synchroniously
		for _, node := range s.config.BlockExporer {
			s.log.Debug("getting data for block-explorer", zap.String("name", node.Name))
			status, err := requestBlockExplorerStats(node.REST)
			if err != nil {
				s.log.Error("Failed to get data for block-explorer", zap.String("node", node.Name), zap.Error(err))
				s.collector.UpdateBlockExplorerStatusAsError(node.Name, err)
			} else {
				s.collector.UpdateBlockExplorerStatus(node.Name, status)
			}
			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Data Node Collector for Prometheus")
			return nil
		case <-ticker.C:
			continue
		}
	}
}
