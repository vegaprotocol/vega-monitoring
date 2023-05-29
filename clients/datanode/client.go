package datanode

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type DataNodeClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewDataNodeClient(apiURL string) *DataNodeClient {
	return &DataNodeClient{
		apiURL:      apiURL,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
