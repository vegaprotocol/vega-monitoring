package nodescanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

func requestBlockExplorerStats(address string) (*types.BlockExplorerStatus, error) {
	// Get core stats
	coreStatus, _, err := requestCoreStats(address, []string{})
	if err != nil {
		return nil, err
	}

	// Get Block Explorer info
	reqURL, err := url.JoinPath(address, "rest", "info")
	if err != nil {
		return nil, fmt.Errorf("failed to get /rest/info of %s, failed to create request URL, %w", address, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get /rest/info %s, failed to create request, %w", address, err)
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to get /rest/info %s, request failed, %w", address, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get /rest/info %s, response status code %d, %w", address, resp.StatusCode, err)
	}

	var payload struct {
		Version     string `json:"version"`
		VersionHash string `json:"commitHash"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to get /rest/info %s, failed to parse response, %w", address, err)
	}

	return &types.BlockExplorerStatus{
		CoreStatus:               *coreStatus,
		BlockExplorerVersion:     payload.Version,
		BlockExplorerVersionHash: payload.VersionHash,
	}, nil
}
