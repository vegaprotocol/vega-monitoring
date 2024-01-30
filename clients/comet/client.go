package comet

import (
	"net/http"
	"time"

	vega_entities "code.vegaprotocol.io/vega/datanode/entities"
	"github.com/vegaprotocol/vega-monitoring/config"
	"golang.org/x/time/rate"
)

type CometClient struct {
	httpClient         *http.Client
	config             *config.CometBFTConfig
	rateLimiter        *rate.Limiter
	validatorByAddress map[string]ValidatorData // local cache
}

func NewCometClient(config *config.CometBFTConfig) *CometClient {
	return &CometClient{
		config:      config,
		rateLimiter: rate.NewLimiter(rate.Every(50*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		validatorByAddress: map[string]ValidatorData{},
	}
}

type ValidatorData struct {
	Address  string
	TmPubKey vega_entities.TendermintPublicKey
}
