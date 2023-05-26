package binance

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type BinanceClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewBinanceClient(apiURL string) *BinanceClient {
	return &BinanceClient{
		apiURL:      apiURL,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
