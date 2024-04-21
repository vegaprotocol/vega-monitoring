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
	"github.com/vegaprotocol/vega-monitoring/internal/retry"
	"github.com/vegaprotocol/vega-monitoring/metamonitoring"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
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

type ethNodeMonitoring struct {
	ethClient *ethutils.EthClient

	nodeName    string
	chainId     string
	networkId   string
	rpcEndpoint string
}

func (s *EthereumMonitoringService) Start(ctx context.Context, statusPublisher metamonitoring.MonitoringStatusPublisher) error {
	var monitoringWg sync.WaitGroup

	svcContext, cancel := context.WithCancel(ctx)
	defer cancel()

	var failure atomic.Bool
	failure.Store(false)

	nodeClients := []ethNodeMonitoring{}

	for idx, chainConfig := range s.cfg {
		if len(chainConfig.RPCEndpoint) < 1 {
			s.logger.Errorf("failed to start the prometheus ethereum monitoring service for network id %s: empty rpc address", chainConfig.NetworkId)
			continue
		}

		ethClient, err := ethutils.NewEthClient(chainConfig.RPCEndpoint, s.logger)
		if err != nil {
			s.logger.Errorf("failed to create ethereum client in the prometheus ethereum monitoring service for network id %s: %s", chainConfig.NetworkId, err.Error())
			continue
		}

		nodeClients = append(nodeClients, ethNodeMonitoring{
			ethClient: ethClient,

			nodeName:    chainConfig.NodeName,
			chainId:     chainConfig.ChainId,
			networkId:   chainConfig.NetworkId,
			rpcEndpoint: chainConfig.RPCEndpoint,
		})

		if len(chainConfig.Accounts) > 0 {
			monitoringWg.Add(1)
			go func(failure *atomic.Bool, callCfg config.EthereumChain) {
				defer monitoringWg.Done()
				if err := s.monitorAccountBalances(
					svcContext,
					ethClient,
					chainConfig.NodeName,
					callCfg.ChainId,
					callCfg.NetworkId,
					callCfg.Accounts,
					callCfg.Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring account balances in the prometheus ethereum monitoring for network id %s: %s", chainConfig.NetworkId, err.Error())
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
					chainConfig.NodeName,
					callCfg.ChainId,
					callCfg.NetworkId,
					callCfg.Calls,
					callCfg.Period,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring ethereum calls in the prometheus ethereum monitoring for network id %s: %s", callCfg.NetworkId, err.Error())
					cancel()
				}
			}(&failure, s.cfg[idx])
		}

		if len(chainConfig.Events) > 0 {
			monitoringWg.Add(1)
			go func(failure *atomic.Bool, callCfg config.EthereumChain) {
				defer monitoringWg.Done()
				if err := s.monitorContractEvents(
					svcContext,
					ethClient,
					chainConfig.NodeName,
					callCfg.Period,
					callCfg.NetworkId,
					callCfg.Events,
				); err != nil {
					failure.Store(true)
					s.logger.Errorf("failed to start monitoring ethereum events for network id %s: %s", callCfg.NetworkId, err.Error())
				}
			}(&failure, s.cfg[idx])
		}
	}

	monitoringWg.Add(1)
	go func(failure *atomic.Bool) {
		defer monitoringWg.Done()

		// We pass Period to each endpoint but here We ignore it
		if err := s.monitorNodeStatuses(svcContext, nodeClients, time.Minute); err != nil {
			s.logger.Errorf("failed to monitor node statuses: %s", err.Error())
			failure.Store(true)
			cancel()
		}
	}(&failure)

	monitoringWg.Add(1)
	go func(failure *atomic.Bool) {
		defer monitoringWg.Done()

		// We pass Period to each endpoint but here We ignore it
		if err := s.monitorNodeHeight(svcContext, nodeClients, time.Minute); err != nil {
			s.logger.Errorf("failed to monitor nodes height: %s", err.Error())
			failure.Store(true)
			cancel()
		}
	}(&failure)

	monitoringWg.Add(1)
	go func(failure *atomic.Bool) {
		defer monitoringWg.Done()
		if err := s.reportState(svcContext, time.Minute, statusPublisher); err != nil {
			s.logger.Errorf("failed to report statuses: %s", err.Error())
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
	nodeName string,
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
			res, err := retry.RetryReturn(3, 2*time.Second, func() (interface{}, error) {
				callCtx, cancel := context.WithTimeout(ctx, defaultCallTimeout)
				defer cancel()
				return ethClient.Call(callCtx, call)
			})

			if err != nil {
				s.logger.Errorf("failed to call ethereum smart contract for ID %s: %s", call.ID(), err.Error())
				s.reportHealth(false, entities.ReasonEthereumContractCallFailure)
				continue
			}

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

			s.collector.UpdateEthereumCallResponse(nodeName, call.ID(), call.ContractAddress().String(), call.MethodName(), float64Res)
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
	nodeName string,
	chainId string,
	networkId string,
	accounts []string,
	period time.Duration,
) error {
	time.Sleep(13 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		// balances := map[string]float64{}

		for _, accAddress := range accounts {
			balance, err := retry.RetryReturn(3, 2*time.Second, func() (float64, error) {
				callCtx, cancel := context.WithTimeout(ctx, defaultCallTimeout)
				defer cancel()

				return ethClient.BalanceWithoutZerosAt(callCtx, common.HexToAddress(accAddress))
			})

			if err != nil {
				s.logger.Errorf("failed to get balance for account %s: %s", accAddress, err.Error())
				s.reportHealth(false, entities.ReasonEthereumGetBalancesFailure)
				continue
			}
			s.collector.UpdateEthereumAccountBalance(nodeName, accAddress, chainId, networkId, balance)
			// fmt.Printf("Balance: %f\n", balance)
			// balances[accAddress] = balance
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

func (s *EthereumMonitoringService) monitorContractEvents(
	ctx context.Context,
	ethClient *ethutils.EthClient,
	nodeName string,
	period time.Duration,
	networkId string,
	cfg []config.EthEvents,
) error {
	time.Sleep(18 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	// TODO Create EventCounts
	eventsCounters := make([]*ethutils.EventsCounter, len(cfg))
	var err error

	for idx, configItem := range cfg {
		eventsCounters[idx], err = ethutils.NewEventsCounterFromConfig(configItem)
		if err != nil {
			return fmt.Errorf("failed to create events counter from config for %s: %w", configItem.Name, err)
		}
	}

	for {

		for _, event := range eventsCounters {
			err := retry.RetryRun(3, 2*time.Second, func() error {
				counterCallCtx, cancel := context.WithTimeout(ctx, defaultCallTimeout)
				defer cancel()

				return event.CallFilterLogs(counterCallCtx, ethClient)
			})

			if err != nil {
				s.logger.Errorf("Failed to filter logs for the events counter(%s): %s", event.Name(), err.Error())
			}
		}

		metrics := []types.EthereumContractsEvents{}
		// Publish metrics to the prometheus
		for _, event := range eventsCounters {
			eventsCalls := event.Count()
			for eventName, count := range eventsCalls {
				metrics = append(metrics, types.EthereumContractsEvents{
					NodeName:        nodeName,
					ID:              event.Name(),
					EventName:       eventName,
					ContractAddress: event.ContractAddress(),
					Count:           count,
				})
			}
		}

		// We need to submit updates for all contracts at the same time.
		s.collector.UpdateEthereumContractEvents(metrics)

		ticker.Reset(period)
		s.reportHealth(true, entities.ReasonUnknown)
		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping filtering events for network id: %s", networkId)
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *EthereumMonitoringService) monitorNodeStatuses(
	ctx context.Context,
	nodes []ethNodeMonitoring,
	period time.Duration,
) error {
	time.Sleep(18 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		statuses := []types.EthereumNodeStatus{}

		for _, node := range nodes {
			if node.ethClient == nil {
				continue
			}

			calCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			nodeReady, err := node.ethClient.Ready(calCtx)
			cancel()
			if err != nil {
				s.logger.Debug(
					fmt.Sprintf("failed to check if node is ready for %s(%s)", node.nodeName, node.rpcEndpoint),
					zap.Error(err),
				)
			}

			statuses = append(statuses, types.EthereumNodeStatus{
				ChainId:     node.chainId,
				NodeName:    node.nodeName,
				RPCEndpoint: node.rpcEndpoint,
				Healthy:     nodeReady,
				UpdateTime:  time.Now(),
			})
		}

		if len(statuses) > 0 {
			s.collector.UpdateEthereumNodeStatuses(statuses)
		}

		ticker.Reset(period)
		s.reportHealth(true, entities.ReasonUnknown)
		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping ethereum node statuses monitoring")
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *EthereumMonitoringService) monitorNodeHeight(
	ctx context.Context,
	nodes []ethNodeMonitoring,
	period time.Duration,
) error {
	time.Sleep(28 * time.Second)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		heights := []types.EthereumNodeHeight{}

		for _, node := range nodes {
			if node.ethClient == nil {
				continue
			}

			calCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			nodeHeight, err := node.ethClient.Height(calCtx)
			cancel()
			if err != nil {
				nodeHeight = 0
				s.logger.Debug(
					fmt.Sprintf("failed to check node height for %s(%s)", node.nodeName, node.rpcEndpoint),
					zap.Error(err),
				)
			}

			heights = append(heights, types.EthereumNodeHeight{
				ChainId:     node.chainId,
				NodeName:    node.nodeName,
				RPCEndpoint: node.rpcEndpoint,
				Height:      nodeHeight,
				UpdateTime:  time.Now(),
			})
		}

		if len(heights) > 0 {
			s.collector.UpdateEthereumNodeHeights(heights)
		}

		ticker.Reset(period)
		s.reportHealth(true, entities.ReasonUnknown)
		select {
		case <-ctx.Done():
			s.logger.Infof("Stopping ethereum node statuses monitoring")
			return nil
		case <-ticker.C:
			continue
		}
	}
}
