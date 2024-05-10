package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type blockResultsResponse struct {
	Result struct {
		Height     string                   `json:"height"`
		TxsResults []blockResultsResponseTx `json:"txs_results"`
	} `json:"result"`
}

type blockResultsResponseTx struct {
	Code      int                           `json:"code"`
	Data      string                        `json:"data"`
	Info      string                        `json:"info"`
	Events    []blockResultsResponseTxEvent `json:"events"`
	Codespace string                        `json:"codespace"`
	GasWanted string                        `json:"gas_wated"`
	GasUsed   string                        `json:"gas_used"`
	Log       string                        `json:"log"`
}

type blockResultsResponseTxEvent struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Index bool   `json:"index"`
	} `json:"attributes"`
}

func (c *CometClient) requestBlockResults(ctx context.Context, block int64) (blockResultsResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return blockResultsResponse{}, fmt.Errorf("failed rate limiter for get block results for block: %d. %w", block, err)
	}

	url := fmt.Sprintf("%s/block_results", c.config.ApiURL)
	if block > 0 {
		url = fmt.Sprintf("%s/block_results?height=%d", c.config.ApiURL, block)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return blockResultsResponse{}, fmt.Errorf("faailed to create request for %s: %w", url, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return blockResultsResponse{}, fmt.Errorf("failed to get block results for block: %d. %w", block, err)
	}
	defer resp.Body.Close()

	var payload blockResultsResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return blockResultsResponse{}, fmt.Errorf("failed to parse response for get block results for block: %d. %w", block, err)
	}

	return payload, nil
}

func (c *CometClient) requestBlockResultsRange(ctx context.Context, startBlock int64, endBlock int64) ([]blockResultsResponse, error) {
	var mut sync.Mutex
	result := []blockResultsResponse{}
	var merr *multierror.Error

	var wg sync.WaitGroup
	for block := startBlock; block <= endBlock; block++ {
		wg.Add(1)
		go func(block int64) {
			defer wg.Done()
			response, err := c.requestBlockResults(ctx, block)

			mut.Lock()
			defer mut.Unlock()

			if err != nil {
				merr = multierror.Append(merr, fmt.Errorf("failed to request blocks results for block %d: %w", block, err))
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
