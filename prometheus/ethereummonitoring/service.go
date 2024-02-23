package ethereummonitoring

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
)

const defaultCallTimeout = 10 * time.Second

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

	svcContext, cancel := context.WithCancel(ctx)
	defer cancel()

	var failure atomic.Bool
	failure.Store(false)

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
			go func(failure *atomic.Bool) {
				defer monitoringWg.Done()
				if err := s.monitorAccountBalances(
					svcContext,
					ethClient,
					chainConfig.ChainId,
					chainConfig.NetworkId,
					s.cfg[idx].Accounts,
					s.cfg[idx].Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring account balances in the prometheus ethereum monitoring for network id %d: %w", chainConfig.NetworkId, err)
					cancel()
				}
			}(&failure)
		}

		if len(chainConfig.Calls) > 0 {
			monitoringWg.Add(1)
			go func(failure *atomic.Bool) {
				defer monitoringWg.Done()
				if err := s.monitorCalls(
					svcContext,
					ethClient,
					chainConfig.ChainId,
					chainConfig.NetworkId,
					s.cfg[idx].Calls,
					s.cfg[idx].Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring ethereum calls in the prometheus ethereum monitoring for network id %d: %w", chainConfig.NetworkId, err)
					cancel()
				}
			}(&failure)
		}
	}

	monitoringWg.Wait()

	if failure.Load() {
		return fmt.Errorf("failed to run ethereum monitoring service: see the error logs above")
	}

	return nil
}

func (s *EthereumMonitoringService) monitorCalls(
	ctx context.Context,
	ethClient *ethutils.EthClient,
	chainId string,
	networkId string,
	cfg []config.EthCall,
	period time.Duration,
) error {
	time.Sleep(11 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	calls := []*ethutils.EthCall{}
	for idx := range cfg {
		call, err := ethutils.NewEthCallFromConfig(cfg[idx])
		if err != nil {
			return fmt.Errorf("failed to create eth call for network id %s: %w", networkId, err)
		}

		calls = append(calls, call)
	}

	for {
		for _, call := range calls {
			callCtx, cancel := context.WithTimeout(ctx, defaultCallTimeout)
			res, err := ethClient.Call(callCtx, call)
			if err != nil {
				s.logger.Errorf("failed to call ethereum smart contract for network %s: %w", err)
				cancel()
				continue
			}
			cancel()
			float64Res, ok := res.(float64)
			if !ok {
				s.logger.Errorf(
					"result for the contract(%s) call (%s) did not return float value. Use transform function that provide float64 response: float64 expected, %#t got",
					call.ContractAddress().String(),
					call.ID(),
					res,
				)
				continue
			}

			s.collector.UpdateEthereumCallResponse(call.ID(), call.ContractAddress().String(), call.MethodName(), float64Res)
		}

		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping eth calls scan for network id: %s", networkId)
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *EthereumMonitoringService) monitorAccountBalances(
	ctx context.Context,
	ethClient *ethutils.EthClient,
	chainId string,
	networkId string,
	accounts []string,
	period time.Duration,
) error {
	time.Sleep(13 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		balances := map[string]float64{}

		for _, accAddress := range accounts {
			callCtx, cancel := context.WithTimeout(ctx, defaultCallTimeout)
			balance, err := ethClient.BalanceWithoutZerosAt(callCtx, common.HexToAddress(accAddress))
			if err != nil {
				cancel()
				s.logger.Errorf("failed to get balance for account %s: %w", accAddress, err)
				continue
			}
			cancel()
			s.collector.UpdateEthereumAccountBalance(accAddress, chainId, networkId, balance)
			// fmt.Printf("Balance: %f\n", balance)
			balances[accAddress] = balance
		}

		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping account scan for network id: %s", networkId)
			return nil
		case <-ticker.C:
			continue
		}
	}
}
