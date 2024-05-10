package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type validatorsResponse struct {
	Result struct {
		BlockHeight string `json:"block_height"`
		Count       string `json:"count"`
		Total       string `json:"total"`
		Validators  []struct {
			Address          string `json:"address"`
			VotingPower      string `json:"voting_power"`
			ProposerPriority string `json:"proposer_priority"`
			PubKey           struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
		} `json:"validators"`
	} `json:"result"`
}

func (c *CometClient) requestValidators(ctx context.Context, block int64) (validatorsResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return validatorsResponse{}, fmt.Errorf("Failed rate limiter for Get Validators for block: %d. %w", block, err)
	}
	url := fmt.Sprintf("%s/validators", c.config.ApiURL)
	if block > 0 {
		url = fmt.Sprintf("%s/validators?height=%d", c.config.ApiURL, block)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return validatorsResponse{}, fmt.Errorf("failed to create validators request for %s url: %w", url, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return validatorsResponse{}, fmt.Errorf("Failed to Get Validators for block: %d. %w", block, err)
	}
	defer resp.Body.Close()
	var payload validatorsResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return validatorsResponse{}, fmt.Errorf("Failed to parse response for Get Validators for block: %d. %w", block, err)
	}

	return payload, nil
}
