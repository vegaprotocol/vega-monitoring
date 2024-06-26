package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vegaprotocol/vega-monitoring/entities"
)

func (c *Client) GetStatistics(ctx context.Context) (*entities.Statistics, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.Join(errWaitingForRateLimiter, fmt.Errorf("failed to get network history segments for %s: %w", c.apiURL, err))
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(statisticsURL, c.apiURL),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request with context: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do http request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code for statistics request: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var payload struct {
		Statistics entities.Statistics `json:"statistics"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics response: %w", err)
	}

	return &payload.Statistics, nil
}
