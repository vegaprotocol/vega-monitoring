package comet

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type CometClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
}

func NewCometClient(apiURL string) *CometClient {
	return &CometClient{
		apiURL:      apiURL,
		rateLimiter: rate.NewLimiter(rate.Every(25*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}
