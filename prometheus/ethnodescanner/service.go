package ethnodescanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
	"go.uber.org/zap"
)

type EthNodeScannerService struct {
	config    []config.EthereumNodeConfig
	collector *collectors.VegaMonitoringCollector
	log       *logging.Logger

	ethMutex   sync.Mutex
	ethClients map[string]*ethutils.EthClient
}

func NewEthNodeScannerService(
	config []config.EthereumNodeConfig,
	collector *collectors.VegaMonitoringCollector,
	log *logging.Logger,
) *EthNodeScannerService {
	log = log.With(zap.String("service", "eth-node-scanner"))

	return &EthNodeScannerService{
		config:     config,
		collector:  collector,
		log:        log,
		ethClients: map[string]*ethutils.EthClient{},
	}
}

func (s *EthNodeScannerService) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	s.initEthereumClients()

	for {
		updateTime := time.Now()
		// TODO(Fixme): We should pass eth client there, We should not create it every time
		statuses := ethutils.CheckETHEndpointList(ctx, s.log, s.config)
		s.collector.UpdateEthereumNodeStatuses(statuses, updateTime)
		s.log.Debug("successfully updated ethereum nodes statuses in prometheus")

		s.log.Debug("getting height for ethereum clients")
		heights := s.getEthNodesHeight(ctx)
		s.collector.UpdateEthereumNodeHeights(heights)
		s.log.Debug("successfully updated heights for ethereum clients in prometheus")

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Ethereum Node Scanner for Prometheus")
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *EthNodeScannerService) initEthereumClients() error {
	var err error
	for _, item := range s.config {
		if _, ok := s.ethClients[item.RPCEndpoint]; ok {
			continue
		}

		s.ethClients[item.RPCEndpoint], err = ethutils.NewEthClient(item.RPCEndpoint, s.log)
		if err != nil {
			return fmt.Errorf("failed to create eth client for the %s endpoint: %w", item.RPCEndpoint, err)
		}
	}

	return nil
}

func (s *EthNodeScannerService) getEthNodesHeight(ctx context.Context) []types.EthereumNodeHeight {
	result := []types.EthereumNodeHeight{}
	for _, item := range s.config {
		client, ok := s.ethClients[item.RPCEndpoint]
		if !ok {
			s.log.Errorf("missing eth client for %s eth endpoint", item.RPCEndpoint)
			continue
		}

		res, err := client.Height(ctx)
		if err != nil {
			s.log.Errorf("failed to get height for %s eth endpoint: %w", item.RPCEndpoint, err)
			continue
		}

		result = append(result, types.EthereumNodeHeight{
			RPCEndpoint: item.RPCEndpoint,
			Name:        item.Name,
			Height:      res,
			UpdateTime:  time.Now(),
		})
	}

	return result
}
