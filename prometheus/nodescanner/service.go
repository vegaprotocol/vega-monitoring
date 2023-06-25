package nodescanner

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus"
	"go.uber.org/zap"
)

type NodeScannerService struct {
	config  *config.MonitoringConfig
	metrics *prometheus.Metrics
	log     *logging.Logger
}

func NewNodeScannerService(
	config *config.MonitoringConfig,
	metrics *prometheus.Metrics,
	log *logging.Logger,
) *NodeScannerService {
	log = log.With(zap.String("service", "node-scanner"))

	return &NodeScannerService{
		config:  config,
		metrics: metrics,
		log:     log,
	}
}

func (s *NodeScannerService) Start(ctx context.Context) error {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		// Go DataNode one-by-one synchroniusly
		for _, node := range s.config.DataNode {
			s.log.Debug("getting data for data-node", zap.String("name", node.Name))
			checkResults, err := requestDataNodeStats(node.REST)
			if err != nil {
				s.log.Error("Failed to get data for", zap.String("node", node.Name), zap.Error(err))
				s.metrics.UpdateDataNodeAsError(node.Name, err)
			} else {
				checkResults.RESTReqDuration, _ = checkREST(node.REST)
				checkResults.GQLReqDuration, _ = checkGQL(node.GraphQL)
				checkResults.GRPCReqDuration, _ = checkGRPC(node.GRPC)
				s.metrics.UpdateDataNodeCheckResults(node.Name, checkResults)
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
