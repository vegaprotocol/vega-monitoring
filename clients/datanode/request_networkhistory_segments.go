package datanode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

func (c *DataNodeClient) requestNetworkHistorySegmets() (networkHistorySegmentsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // TODO: Pass parent context
	defer cancel()

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return networkHistorySegmentsResponse{}, fmt.Errorf("Failed rate limiter for Get Network History Segmets for %s. %w", c.apiURL, err)
	}
	url := fmt.Sprintf("%s/api/v2/networkhistory/segments", c.apiURL)
	resp, err := http.Get(url)
	if err != nil {
		return networkHistorySegmentsResponse{}, fmt.Errorf("Failed to Get Network History Segmets for  %s. %w", c.apiURL, err)
	}
	defer resp.Body.Close()
	var payload networkHistorySegmentsResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return networkHistorySegmentsResponse{}, fmt.Errorf("Failed to parse response for Get Network History Segmets for  %s. %w", c.apiURL, err)
	}

	return payload, nil
}
