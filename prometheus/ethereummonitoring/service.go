package ethereummonitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
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
	var monitoringWg sync.WaitGroup

	for idx, chainConfig := range s.cfg {
		if len(chainConfig.RPCEndpoint) < 1 {
			s.logger.Errorf("failed to start the prometheus ethereum monitoring service for network id %d: empty rpc address", chainConfig.NetworkId)
			continue
		}

		ethClient, err := ethutils.NewEthClient(chainConfig.RPCEndpoint, s.logger)
		if err != nil {
			s.logger.Errorf("failed to create ethereum client in the prometheus ethereum monitoring service for network id %d: %w", chainConfig.NetworkId, err)
			continue
		}
		if len(chainConfig.Accounts) > 0 {
			monitoringWg.Add(1)
			go func() {
				defer monitoringWg.Done()
				if err := s.monitorAccountBalances(ctx, ethClient, chainConfig.ChainId, chainConfig.NetworkId, s.cfg[idx].Accounts); err != nil {
					s.logger.Errorf("failed to start monitoring account balances in the prometheus ethereum monitoring for network id %d: %w", chainConfig.NetworkId, err)
				}
			}()
		}
	}

	monitoringWg.Wait()

	return nil
}

func (s *EthereumMonitoringService) monitorAccountBalances(ctx context.Context, ethClient *ethutils.EthClient, chainId, networkId string, accounts []string) error {

	time.Sleep(25 * time.Second)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		balances := map[string]float64{}

		for _, accAddress := range accounts {
			callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			balance, err := ethClient.BalanceWithoutZerosAt(callCtx, common.HexToAddress(accAddress))
			if err != nil {
				cancel()
				s.logger.Errorf("failed to get balance for account %s: %w", accAddress, err)
				continue
			}
			cancel()
			s.collector.UpdateEthereumAccountBalance(accAddress, chainId, networkId, balance)
			fmt.Printf("Balance: %f\n", balance)
			balances[accAddress] = balance
		}

		select {
		case <-ctx.Done():
			s.logger.Info("Stopping account scan")
			return nil
		case <-ticker.C:
			continue
		}
	}
}
