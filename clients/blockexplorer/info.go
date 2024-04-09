package blockexplorer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	explorerapi "code.vegaprotocol.io/vega/protos/blockexplorer/api/v1"
)

func (c *Client) GetInfo(ctx context.Context) (*explorerapi.InfoResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.Join(errWaitingForRateLimiter, fmt.Errorf("failed to get network history segments for %s: %w", c.apiURL, err))
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(infoURL, c.apiURL),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request with context: %w", err)
	}
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to do a request for block explorer info: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code for block explorer info request: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var payload explorerapi.InfoResponse
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &payload, nil
}
