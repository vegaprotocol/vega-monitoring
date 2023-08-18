package ethnodescanner

import (
	"context"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"go.uber.org/zap"
)

type EthNodeScannerService struct {
	config    []config.EthereumNodeConfig
	collector *collectors.VegaMonitoringCollector
	log       *logging.Logger
}

func NewEthNodeScannerService(
	config []config.EthereumNodeConfig,
	collector *collectors.VegaMonitoringCollector,
	log *logging.Logger,
) *EthNodeScannerService {
	log = log.With(zap.String("service", "eth-node-scanner"))

	return &EthNodeScannerService{
		config:    config,
		collector: collector,
		log:       log,
	}
}

func (s *EthNodeScannerService) Start(ctx context.Context) error {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		updateTime := time.Now()
		statuses := ethutils.CheckETHEndpointList(ctx, s.log, s.config)

		s.collector.UpdateEthereumNodeStatuses(statuses, updateTime)

		s.log.Debug("successfully updated Ethereum Nodes statuses in prometheus")

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Ethereum Node Scanner for Prometheus")
			return nil
		case <-ticker.C:
			continue
		}
	}
}
