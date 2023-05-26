package coingecko

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type CoingeckoClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewCoingeckoClient(apiURL string) *CoingeckoClient {
	return &CoingeckoClient{
		apiURL:      apiURL,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
