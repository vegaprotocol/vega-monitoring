package datanode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	api "code.vegaprotocol.io/protos/vega/api/v1"
)

func (c *DataNodeClient) GetStatistics(ctx context.Context) (*api.Statistics, error) {
	resp, _, err := c.GetStatisticsWithHeaders(ctx, nil)

	return resp, err
}

func (c *DataNodeClient) GetStatisticsWithHeaders(ctx context.Context, wantedHeaders []string) (*api.Statistics, map[string]string, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, nil, errors.Join(errWaitingForRateLimiter, fmt.Errorf("failed to get network history segments for %s: %w", c.apiURL, err))
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(statisticsURL, c.apiURL),
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new request with context: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to do http request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("invalid response code for statistics request: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var payload struct {
		Statistics api.Statistics `json:"statistics"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal statistics response: %w", err)
	}

	headerValues := map[string]string{}

	for _, header := range wantedHeaders {
		headerValues[header] = resp.Header.Get(header)
	}

	return &payload.Statistics, headerValues, nil
}
