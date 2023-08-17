package ethutils

import (
	"context"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type healthcheckResult struct {
	endpoint string
	result   bool
}

func CheckETHEndpointList(ctx context.Context, log *logging.Logger, endpoints []string) map[string]bool {
	var wg sync.WaitGroup
	ch := make(chan healthcheckResult, len(endpoints))
	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()
			singleResult := CheckETHEndpoint(ctx, log, endpoint)
			ch <- healthcheckResult{
				endpoint: endpoint,
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

func CheckETHEndpoint(ctx context.Context, log *logging.Logger, endpoint string) bool {
	client, err := ethclient.DialContext(ctx, endpoint)
	if err != nil {
		log.Debug("Failed to create ETH Client", zap.String("endpoint", endpoint), zap.Error(err))
		return false
	}
	latestBlock, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Debug("Failed to get latest Eth block", zap.String("endpoint", endpoint), zap.Error(err))
		return false
	}
	latestBlockTime := time.Unix(int64(latestBlock.Time), 0)
	currentTime := time.Now()
	timeDiff := currentTime.Sub(latestBlockTime)
	if timeDiff > 2*time.Minute {
		log.Debug("Latest Eth block is far behind", zap.String("endpoint", endpoint), zap.Duration("time behind", timeDiff))
		return false
	}
	log.Debug("Node is healthy", zap.String("endpoint", endpoint), zap.Duration("time behind", timeDiff))
	return true
}
