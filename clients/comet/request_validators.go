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

func (c *CometClient) requestValidators(block int64) (validatorsResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return validatorsResponse{}, fmt.Errorf("Failed rate limiter for Get Validators for block: %d. %w", block, err)
	}
	url := fmt.Sprintf("%s/validators", c.config.CometURL)
	if block > 0 {
		url = fmt.Sprintf("%s/validators?height=%d", c.config.CometURL, block)
	}
	resp, err := http.Get(url)
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
