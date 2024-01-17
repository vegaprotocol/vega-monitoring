package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type statusResponseRaw struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHash   string `json:"latest_block_hash"`
			LatestAppHash     string `json:"latest_app_hash"`
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`

			EarliestBlockHash   string `json:"earliest_block_hash"`
			EarliestAppHash     string `json:"earliest_app_hash"`
			EarliestBlockHeight string `json:"earliest_block_height"`
			EarliestBlockTime   string `json:"earliest_block_time"`
			CatchingUp          bool   `json:"catching_up"`
		} `json:"sync_info"`
	} `json:"result"`
}

type StatusSyncInfo struct {
	LatestBlockHash   string    `json:"latest_block_hash"`
	LatestAppHash     string    `json:"latest_app_hash"`
	LatestBlockHeight int64     `json:"latest_block_height"`
	LatestBlockTime   time.Time `json:"latest_block_time"`

	EarliestBlockHash   string    `json:"earliest_block_hash"`
	EarliestAppHash     string    `json:"earliest_app_hash"`
	EarliestBlockHeight int64     `json:"earliest_block_height"`
	EarliestBlockTime   time.Time `json:"earliest_block_time"`
	CatchingUp          bool      `json:"catching_up"`
}

type StatusResponse struct {
	SyncInfo StatusSyncInfo `json:"sync_info"`
}

func (c *CometClient) Status(ctx context.Context) (*StatusResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter failed to wait for get status: %w", err)
	}

	url := fmt.Sprintf("%s/status", c.config.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to response for the /status comet endpoint:%w", err)
	}
	defer resp.Body.Close()

	var payload statusResponseRaw
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse response for the /status comet endpoint: %w", err)
	}

	latestBlockHeight, err := strconv.ParseInt(payload.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest block height for the /status comet endpoint: %w", err)
	}
	latestBlockTime, err := time.Parse(time.RFC3339, payload.Result.SyncInfo.LatestBlockTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest block time for the /status comet endpoint: %w", err)
	}

	earliestBlockHeight, err := strconv.ParseInt(payload.Result.SyncInfo.EarliestBlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse earliest block height for the /status comet endpoint: %w", err)
	}
	earliestBlockTime, err := time.Parse(time.RFC3339, payload.Result.SyncInfo.EarliestBlockTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse earliest block time for the /status comet endpoint: %w", err)
	}

	return &StatusResponse{
		SyncInfo: StatusSyncInfo{
			LatestBlockHash:   payload.Result.SyncInfo.LatestBlockHash,
			LatestAppHash:     payload.Result.SyncInfo.LatestAppHash,
			LatestBlockHeight: latestBlockHeight,
			LatestBlockTime:   latestBlockTime,

			EarliestBlockHash:   payload.Result.SyncInfo.EarliestBlockHash,
			EarliestAppHash:     payload.Result.SyncInfo.EarliestAppHash,
			EarliestBlockHeight: earliestBlockHeight,
			EarliestBlockTime:   earliestBlockTime,
			CatchingUp:          payload.Result.SyncInfo.CatchingUp,
		},
	}, nil
}

func (c *CometClient) EarliestBlockHeight(ctx context.Context) (int64, error) {
	res, err := c.Status(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get the /status response from the comet api: %w", err)
	}

	return res.SyncInfo.EarliestBlockHeight, nil
}

func (c *CometClient) LatestLocalBlockHeight(ctx context.Context) (int64, error) {
	res, err := c.Status(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get the /status response from the comet api: %w", err)
	}

	return res.SyncInfo.LatestBlockHeight, nil
}
