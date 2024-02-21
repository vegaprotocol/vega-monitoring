package ethereummonitoring

import (
	"context"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
)

type EthereumMonitoringService struct {
	cfg       []config.EthereumChain
	collector *collectors.VegaMonitoringCollector
	logger    *logging.Logger
}

func NewEthereumMonitoringService(
	cfg []config.EthereumChain,
	collector *collectors.VegaMonitoringCollector,
	logger *logging.Logger,
) *EthereumMonitoringService {
	return &EthereumMonitoringService{
		cfg:       cfg,
		collector: collector,
		logger:    logger,
	}
}

func (s *EthereumMonitoringService) Start(ctx context.Context) error {

	return nil
}
