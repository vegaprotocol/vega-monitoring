package datanode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type NetworkHistorySegment struct {
	Height    int64  `db:"height"`
	SegmentId string `db:"segment_id"`
	DataNode  string `db:"data_node"`
}

func (c *DataNodeClient) GetNetworkHistorySegments(ctx context.Context, fromBlock, toBlock int64) ([]*NetworkHistorySegment, error) {
	response, err := c.requestNetworkHistorySegments(ctx)
	if err != nil {
		return nil, err
	}
	result := []*NetworkHistorySegment{}
	for _, segment := range response.Segments {
		height, err := strconv.ParseInt(segment.ToHeight, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ToHeight %s, %w", segment.ToHeight, err)
		}

		// filter blocks We do not want
		if height < fromBlock || height > toBlock {
			continue
		}

		result = append(result, &NetworkHistorySegment{
			Height:    height,
			SegmentId: segment.SegmentId,
			DataNode:  c.apiURL,
		})
	}

	return result, nil
}

type networkHistorySegmentsResponse struct {
	Segments []struct {
		FromHeight        string `json:"fromHeight"`
		ToHeight          string `json:"toHeight"`
		SegmentId         string `json:"historySegmentId"`
		PreviousSegmentId string `json:"previousHistorySegmentId"`
		DatabaseVersion   string `json:"databaseVersion"`
		ChainId           string `json:"chainId"`
	} `json:"segments"`
}

func (c *DataNodeClient) requestNetworkHistorySegments(ctx context.Context) (networkHistorySegmentsResponse, error) {
	innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.rateLimiter.Wait(innerCtx); err != nil {
		return networkHistorySegmentsResponse{}, errors.Join(errWaitingForRateLimiter, fmt.Errorf("failed to get network history segments for %s: %w", c.apiURL, err))
	}
	url := fmt.Sprintf(networkHistorySegmentURL, c.apiURL)
	resp, err := http.Get(url)
	if err != nil {
		return networkHistorySegmentsResponse{}, fmt.Errorf("failed to call get for network history segments to %s: %w", c.apiURL, err)
	}
	defer resp.Body.Close()
	var payload networkHistorySegmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return networkHistorySegmentsResponse{}, fmt.Errorf("failed to unmarshal http request for %s: %w", c.apiURL, err)
	}

	return payload, nil
}
