package datanode

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var (
	errWaitingForRateLimiter = errors.New("failed waiting for rate limiter")
)

const (
	statisticsURL            = "%s/statistics"
	networkHistorySegmentURL = "%s/api/v2/networkhistory/segments"
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
