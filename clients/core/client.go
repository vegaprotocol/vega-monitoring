package core

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var (
	errWaitingForRateLimiter = errors.New("failed waiting for rate limiter")
)

const statisticsURL = "%s/statistics"

type Client struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewCoreClient(apiRestURL string) *Client {
	return &Client{
		apiURL:      apiRestURL,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
