package datanode

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type VegaClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewVegaClient(apiURL string) *VegaClient {
	return &VegaClient{
		apiURL:      apiURL,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
