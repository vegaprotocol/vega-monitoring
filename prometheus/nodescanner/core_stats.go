package nodescanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vegaprotocol/vega-monitoring/prometheus"
)

func requestCoreStats(address string, headers []string) (*prometheus.CoreCheckResults, map[string]string, error) {
	reqURL, err := url.JoinPath(address, "statistics")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check REST of %s, failed to create request URL, %w", address, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check REST %s, failed to create request, %w", address, err)
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to check REST %s, request failed, %w", address, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to check REST %s, response status code %d, %w", address, resp.StatusCode, err)
	}

	var payload struct {
		Statistics struct {
			BlockHeight        uint64 `json:"blockHeight,string"`
			CurrentTime        string `json:"currentTime"`
			VegaTime           string `json:"vegaTime"`
			CoreChainId        string `json:"chainId"`
			CoreAppVersion     string `json:"appVersion"`
			CoreAppVersionHash string `json:"appVersionHash"`
		} `json:"statistics"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("failed to check REST %s, failed to parse response, %w", address, err)
	}
	currentTime, err := time.Parse(time.RFC3339, payload.Statistics.CurrentTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check REST %s, failed to parse currentTime %s, %w", address, payload.Statistics.CurrentTime, err)
	}
	vegaTime, err := time.Parse(time.RFC3339, payload.Statistics.VegaTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check REST %s, failed to parse vegaTime %s, %w", address, payload.Statistics.VegaTime, err)
	}

	headerValues := map[string]string{}

	for _, header := range headers {
		headerValues[header] = resp.Header.Get(header)
	}

	return &prometheus.CoreCheckResults{
			CurrentTime:        currentTime,
			CoreTime:           vegaTime,
			CoreBlockHeight:    payload.Statistics.BlockHeight,
			CoreChainId:        payload.Statistics.CoreChainId,
			CoreAppVersion:     payload.Statistics.CoreAppVersion,
			CoreAppVersionHash: payload.Statistics.CoreAppVersionHash,
		},
		headerValues,
		nil
}
