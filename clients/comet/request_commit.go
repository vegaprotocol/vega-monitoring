package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type commitResponse struct {
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

func (c *CometClient) requestCommit(ctx context.Context, block int64) (commitResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return commitResponse{}, fmt.Errorf("Failed rate limiter for Get Commit Data for block: %d. %w", block, err)
	}
	url := fmt.Sprintf("%s/commit", c.config.ApiURL)
	if block > 0 {
		url = fmt.Sprintf("%s/commit?height=%d", c.config.ApiURL, block)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return commitResponse{}, fmt.Errorf("failed to create request for url %s: %w", url, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return commitResponse{}, fmt.Errorf("Failed to Get Commit Data for block: %d. %w", block, err)
	}
	defer resp.Body.Close()
	var payload commitResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return commitResponse{}, fmt.Errorf("Failed to parse response for Get Commit Data for block: %d. %w", block, err)
	}

	return payload, nil
}

func (c *CometClient) requestCommitRange(ctx context.Context, startBlock int64, endBlock int64) ([]commitResponse, error) {
	var mut sync.Mutex
	result := []commitResponse{}
	var merr *multierror.Error

	var wg sync.WaitGroup
	for block := startBlock; block <= endBlock; block++ {
		wg.Add(1)
		go func(block int64) {
			defer wg.Done()
			response, err := c.requestCommit(ctx, block)

			mut.Lock()
			defer mut.Unlock()

			if err != nil {
				merr = multierror.Append(merr, fmt.Errorf("failed to request comit range for block %d: %w", block, err))
				return
			}
			result = append(result, response)
		}(block)
	}
	wg.Wait()

	if merr != nil {
		return nil, merr.ErrorOrNil()
	}

	return result, nil
}
