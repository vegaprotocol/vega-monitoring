package ethutils

import (
	"context"
	"math/big"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vegaprotocol/vega-monitoring/config"
	"go.uber.org/zap"
)

type healthcheckResult struct {
	endpoint string
	result   bool
}

func CheckETHEndpointList(ctx context.Context, log *logging.Logger, endpoints []config.EthereumNodeConfig) map[string]bool {
	var wg sync.WaitGroup
	ch := make(chan healthcheckResult, len(endpoints))
	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint config.EthereumNodeConfig) {
			defer wg.Done()
			singleResult := CheckETHEndpoint(ctx, log, endpoint)
			ch <- healthcheckResult{
				endpoint: endpoint.RPCEndpoint,
				result:   singleResult,
			}
		}(endpoint)
	}
	wg.Wait()
	close(ch)
	result := map[string]bool{}
	for singleResult := range ch {
		result[singleResult.endpoint] = singleResult.result
	}
	return result
}

func CheckETHEndpoint(ctx context.Context, log *logging.Logger, endpoint config.EthereumNodeConfig) bool {
	client, err := ethclient.DialContext(ctx, endpoint.RPCEndpoint)
	if err != nil {
		log.Debug("Failed to create ETH Client", zap.String("endpoint", endpoint.RPCEndpoint), zap.Error(err))
		return false
	}
	latestBlock, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Debug("Failed to get latest Eth block", zap.String("endpoint", endpoint.RPCEndpoint), zap.Error(err))
		return false
	}
	latestBlockTime := time.Unix(int64(latestBlock.Time), 0)
	currentTime := time.Now()
	timeDiff := currentTime.Sub(latestBlockTime)
	if timeDiff > 2*time.Minute {
		log.Debug("Latest Eth block is far behind", zap.String("endpoint", endpoint.RPCEndpoint), zap.Duration("time behind", timeDiff))
		return false
	}
	if len(endpoint.VegaCollateralBridgeAddress) > 0 {
		bridgeAddress := common.HexToAddress(endpoint.VegaCollateralBridgeAddress)
		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{bridgeAddress},
			FromBlock: big.NewInt(0).Sub(latestBlock.Number, big.NewInt(600)), // two hours of blocks
		}
		logs, err := client.FilterLogs(ctx, filterQuery)
		if err != nil {
			log.Debug("Failed to eth_getLogs",
				zap.String("endpoint", endpoint.RPCEndpoint),
				zap.String("collateralBridge", endpoint.VegaCollateralBridgeAddress),
				zap.Error(err),
			)
			return false
		}
		log.Debug("Successfully fetched eth_getLogs",
			zap.String("endpoint", endpoint.RPCEndpoint),
			zap.String("collateralBridge", endpoint.VegaCollateralBridgeAddress),
			zap.Int("count", len(logs)),
		)
	}

	log.Debug("Node is healthy", zap.String("endpoint", endpoint.RPCEndpoint), zap.Duration("time behind", timeDiff))
	return true
}
