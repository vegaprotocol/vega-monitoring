package comet

import (
	"net/http"
	"time"

	"github.com/vegaprotocol/data-metrics-store/config"
	"golang.org/x/time/rate"
)

type CometClient struct {
	httpClient  *http.Client
	config      *config.CometBFTConfig
	rateLimiter *rate.Limiter
}

func NewCometClient(config *config.CometBFTConfig) *CometClient {
	return &CometClient{
		config:      config,
		rateLimiter: rate.NewLimiter(rate.Every(10*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
