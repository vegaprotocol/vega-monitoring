package datanode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type statisticsRaw struct {
	Statistics struct {
		BlockHeight string `json:"blockHeight"`
		CurrentTime string `json:"currentTime"`
		VegaTime    string `json:"vegaTime"`
	} `json:"statistics"`
}

type Statistics struct {
	BlockHeight int64
	CurrentTime time.Time
	VegaTime    time.Time
}

func (c *VegaClient) requestStatistics() (*statisticsRaw, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // TODO: Pass parent context
	defer cancel()

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for rate limit on %s: %w", c.apiURL, err)
	}

	url := fmt.Sprintf("%s/statistics", c.apiURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics response for %s: %w", c.apiURL, err)
	}

	defer resp.Body.Close()

	var payload statisticsRaw
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse response for statistics response from %s: %w", c.apiURL, err)
	}

	return &payload, nil
}

func (c *VegaClient) GetStatistics() (*Statistics, error) {
	response, err := c.requestStatistics()
	if err != nil {
		return nil, err
	}

	result := &Statistics{}

	result.BlockHeight, err = strconv.ParseInt(response.Statistics.BlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to int: %w", err)
	}

	result.VegaTime, err = time.Parse(time.RFC3339Nano, response.Statistics.VegaTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vega time: %w", err)
	}

	result.CurrentTime, err = time.Parse(time.RFC3339Nano, response.Statistics.CurrentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current time: %w", err)
	}

	return result, nil
}
