package etherscan

import (
	"fmt"
	"time"

	"github.com/vegaprotocol/data-metrics-store/types"
	"golang.org/x/time/rate"
)

type EtherscanClient struct {
	apiURL      string
	apikey      string
	rateLimiter *rate.Limiter
}

func NewEtherscanClient(
	ethNetwork types.ETHNetwork,
	apikey string,
) (*EtherscanClient, error) {
	var apiURL string
	switch ethNetwork {
	case types.ETHMainnet:
		apiURL = "https://api.etherscan.io/api"
	case types.ETHSepolia:
		apiURL = "https://api-sepolia.etherscan.io/api"
	default:
		return nil, fmt.Errorf("failed to get etherscan client, not supported ethereum network %s", ethNetwork)
	}

	return &EtherscanClient{
		apiURL:      apiURL,
		apikey:      apikey,
		rateLimiter: etherscanRateLimiter(apikey),
	}, nil
}

func etherscanRateLimiter(apikey string) *rate.Limiter {
	// API requests to Etherscan are rate limited
	// - with APIKEY - the rate limiting is 5 requests a second
	// - without APIKEY - the rate limit is 1 request every 5 seconds
	if len(apikey) > 0 {
		return rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	}
	return rate.NewLimiter(rate.Every(5*time.Second), 1)
}
