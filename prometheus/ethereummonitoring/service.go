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
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/metamonitoring"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"go.uber.org/zap"
)

const defaultCallTimeout = 10 * time.Second

type EthereumMonitoringService struct {
	cfg                []config.EthereumChain
	collector          *collectors.VegaMonitoringCollector
	logger             *logging.Logger
	monitoringStatuses []healthStatus
	msLock             sync.Mutex
}

type healthStatus struct {
	healthy bool
	reason  entities.UnhealthyReason
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

		monitoringStatuses: []healthStatus{},
	}
}

func (s *EthereumMonitoringService) Start(ctx context.Context, statusPublisher metamonitoring.MonitoringStatusPublisher) error {
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
			go func(failure *atomic.Bool, callCfg config.EthereumChain) {
				defer monitoringWg.Done()
				if err := s.monitorAccountBalances(
					svcContext,
					ethClient,
					callCfg.ChainId,
					callCfg.NetworkId,
					callCfg.Accounts,
					callCfg.Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring account balances in the prometheus ethereum monitoring for network id %d: %w", chainConfig.NetworkId, err)
					cancel()
				}
			}(&failure, s.cfg[idx])
		}

		if len(chainConfig.Calls) > 0 {
			monitoringWg.Add(1)
			go func(failure *atomic.Bool, callCfg config.EthereumChain) {
				defer monitoringWg.Done()
				if err := s.monitorCalls(
					svcContext,
					ethClient,
					callCfg.ChainId,
					callCfg.NetworkId,
					callCfg.Calls,
					callCfg.Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring ethereum calls in the prometheus ethereum monitoring for network id %d: %w", callCfg.NetworkId, err)
					cancel()
				}
			}(&failure, s.cfg[idx])
		}
	}

	monitoringWg.Add(1)
	go func(failure *atomic.Bool) {
		defer monitoringWg.Done()
		if err := s.reportState(svcContext, time.Minute, statusPublisher); err != nil {
			s.logger.Errorf("failed to report statuses: %w", err)
			failure.Store(true)
			cancel()
		}
	}(&failure)

	monitoringWg.Wait()

	if failure.Load() {
		return fmt.Errorf("failed to run ethereum monitoring service: see the error logs above")
	}

	return nil
}

func (s *EthereumMonitoringService) reportState(ctx context.Context, period time.Duration, statusPublisher metamonitoring.MonitoringStatusPublisher) error {

	time.Sleep(30 * 2)

	ticker := time.NewTicker(period * 2)
	defer ticker.Stop()

	for {
		s.msLock.Lock()
		reportedStatuses := s.monitoringStatuses
		s.monitoringStatuses = []healthStatus{}
		s.msLock.Unlock()

		reported := false
		// Report all unhealthy statuses
		for _, status := range reportedStatuses {
			if !status.healthy {
				if err := statusPublisher.PublishWithReason(false, status.reason); err != nil {
					s.logger.Error("failed to publish failure reason in the EthereumMonitoringService.reportState", zap.Error(err))
				}
				reported = true
			}
		}

		// Not reported status yet, report successful
		if !reported {
			if err := statusPublisher.Publish(true); err != nil {
				s.logger.Error("failed to publish healthy reason in the EthereumMonitoringService.reportState", zap.Error(err))
			}
		}

		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping eth calls status publisher")
			return nil
		case <-ticker.C:
			continue
		}
	}
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
				s.logger.Errorf("failed to call ethereum smart contract for ID %s: %s", call.ID(), err.Error())
				s.reportHealth(false, entities.ReasonEthereumContractCallFailure)
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
				s.reportHealth(false, entities.ReasonEthereumContractInvalidResponseType)
				continue
			}

			s.collector.UpdateEthereumCallResponse(call.ID(), call.ContractAddress().String(), call.MethodName(), float64Res)
		}

		s.reportHealth(true, entities.ReasonUnknown)

		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping eth calls scan for network id: %s", networkId)
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *EthereumMonitoringService) reportHealth(healthy bool, reason entities.UnhealthyReason) {
	s.msLock.Lock()
	defer s.msLock.Unlock()

	s.monitoringStatuses = append(s.monitoringStatuses, healthStatus{
		healthy: healthy,
		reason:  reason,
	})
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
				s.reportHealth(false, entities.ReasonEthereumGetBalancesFailure)
				continue
			}
			cancel()
			s.collector.UpdateEthereumAccountBalance(accAddress, chainId, networkId, balance)
			// fmt.Printf("Balance: %f\n", balance)
			balances[accAddress] = balance
		}

		s.reportHealth(true, entities.ReasonUnknown)
		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping account scan for network id: %s", networkId)
			return nil
		case <-ticker.C:
			continue
		}
	}
}
