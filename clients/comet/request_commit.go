package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type cometCommitResponse struct {
	Result struct {
		SignedHeader struct {
			Header struct {
				Height          string `json:"height"`
				Time            string `json:"time"`
				ProposerAddress string `json:"proposer_address"`
			} `json:"header"`
			Commit struct {
				Height     string `json:"height"`
				Signatures []struct {
					ValidatorAddress string `json:"validator_address"`
					Timestamp        string `json:"timestamp"`
				} `json:"signatures"`
			} `json:"commit"`
		} `json:"signed_header"`
	} `json:"result"`
}

func (c *CometClient) requestCometCommit(block int64) (cometCommitResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return cometCommitResponse{}, fmt.Errorf("Failed rate limiter for Get Commit Data for block: %d. %w", block, err)
	}
	url := fmt.Sprintf("%s/commit", c.apiURL)
	if block > 0 {
		url = fmt.Sprintf("%s/commit?height=%d", c.apiURL, block)
	}
	resp, err := http.Get(url)
	if err != nil {
		return cometCommitResponse{}, fmt.Errorf("Failed to Get Commit Data for block: %d. %w", block, err)
	}
	defer resp.Body.Close()
	var payload cometCommitResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return cometCommitResponse{}, fmt.Errorf("Failed to parse response for Get Commit Data for block: %d. %w", block, err)
	}

	return payload, nil
}

func (c *CometClient) requestCometCommitRange(startBlock int64, endBlock int64) (result []cometCommitResponse, err error) {
	var wg sync.WaitGroup
	ch := make(chan cometCommitResponse, endBlock-startBlock+1)
	for block := startBlock; block <= endBlock; block++ {
		wg.Add(1)
		go func(block int64) {
			defer wg.Done()
			response, err := c.requestCometCommit(block)
			if err != nil {
				fmt.Println(err)
				response = cometCommitResponse{}
				response.Result.SignedHeader.Header.Height = strconv.FormatInt(block, 10)
			}
			ch <- response
		}(block)
	}
	wg.Wait()
	close(ch)
	for response := range ch {
		result = append(result, response)
	}
	return
}
