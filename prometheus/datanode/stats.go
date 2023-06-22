package datanode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/vegaprotocol/vega-monitoring/prometheus"
)

var (
	timeout = 2 * time.Second
)

func requestStats(address string) (*prometheus.DataNodeChecksResults, error) {
	reqURL, err := url.JoinPath(address, "statistics")
	if err != nil {
		return nil, fmt.Errorf("failed to check REST of %s, failed to create request URL, %w", address, err)
	}

	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to create request, %w", address, err)
	}
	resp, err := http.DefaultClient.Do(req)
	_ = time.Since(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, request failed, %w", address, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check REST %s, response status code %d, %w", address, resp.StatusCode, err)
	}

	var payload struct {
		Statistics struct {
			BlockHeight    uint64 `json:"blockHeight,string"`
			CurrentTime    string `json:"currentTime"`
			VegaTime       string `json:"vegaTime"`
			ChainId        string `json:"chainId"`
			AppVersion     string `json:"appVersion"`
			AppVersionHash string `json:"appVersionHash"`
		} `json:"statistics"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse response, %w", address, err)
	}
	currentTime, err := time.Parse(time.RFC3339, payload.Statistics.CurrentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse currentTime %s, %w", address, payload.Statistics.CurrentTime, err)
	}
	vegaTime, err := time.Parse(time.RFC3339, payload.Statistics.VegaTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse vegaTime %s, %w", address, payload.Statistics.VegaTime, err)
	}
	strDataNodeBlockHeight := resp.Header.Get("x-block-height")
	if len(strDataNodeBlockHeight) == 0 {
		return nil, fmt.Errorf("failed to check REST %s, failed to get x-block-height response header, %w", address, err)
	}
	dataNodeBlockHeight, err := strconv.ParseUint(strDataNodeBlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse x-block-height response header %s, %w", address, strDataNodeBlockHeight, err)
	}
	strDataNodeTime := resp.Header.Get("x-block-timestamp")
	if len(strDataNodeTime) == 0 {
		return nil, fmt.Errorf("failed to check REST %s, failed to get x-block-timestamp response header, %w", address, err)
	}
	intDataNodeTime, err := strconv.ParseInt(strDataNodeTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse x-block-timestamp response header to int %s, %w", address, strDataNodeTime, err)
	}
	dataNodeTime := time.Unix(intDataNodeTime, 0)

	return &prometheus.DataNodeChecksResults{
		CurrentTime:         currentTime,
		CoreTime:            vegaTime,
		DataNodeTime:        dataNodeTime,
		CoreBlockHeight:     payload.Statistics.BlockHeight,
		DataNodeBlockHeight: dataNodeBlockHeight,
		ChainId:             payload.Statistics.ChainId,
		AppVersion:          payload.Statistics.AppVersion,
		AppVersionHash:      payload.Statistics.AppVersionHash,
	}, nil
}
