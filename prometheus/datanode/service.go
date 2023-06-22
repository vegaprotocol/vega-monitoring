package datanode

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus"
	"go.uber.org/zap"
)

type DataNodeCheckerService struct {
	config  *[]config.DataNodeConfig
	metrics *prometheus.Metrics
	log     *logging.Logger
}

func NewDataNodeCheckerService(
	config *[]config.DataNodeConfig,
	metrics *prometheus.Metrics,
	log *logging.Logger,
) *DataNodeCheckerService {

	return &DataNodeCheckerService{
		config:  config,
		metrics: metrics,
		log:     log,
	}
}

func (s *DataNodeCheckerService) Start(ctx context.Context) error {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		for _, node := range *s.config {
			s.log.Debug("getting data for data-node", zap.String("name", node.Name))
			restResult, err := checkREST(node.REST)
			if err != nil {
				s.log.Error("Failed to get data for", zap.String("node", node.Name), zap.Error(err))
				s.metrics.UpdateNodeAsError(node.Name, err)
			} else {
				s.metrics.UpdateNodeRESTResults(node.Name, restResult)
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
