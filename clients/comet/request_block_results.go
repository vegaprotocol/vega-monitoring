package comet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
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

func (c *CometClient) requestBlockResults(block int64) (blockResultsResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return blockResultsResponse{}, fmt.Errorf("failed rate limiter for get block results for block: %d. %w", block, err)
	}
	url := fmt.Sprintf("%s/block_results", c.config.ApiURL)
	if block > 0 {
		url = fmt.Sprintf("%s/block_results?height=%d", c.config.ApiURL, block)
	}
	resp, err := http.Get(url)
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

func (c *CometClient) requestBlockResultsRange(startBlock int64, endBlock int64) (result []blockResultsResponse, err error) {
	var wg sync.WaitGroup
	ch := make(chan blockResultsResponse, endBlock-startBlock+1)
	for block := startBlock; block <= endBlock; block++ {
		wg.Add(1)
		go func(block int64) {
			defer wg.Done()
			response, err := c.requestBlockResults(block)
			if err != nil {
				fmt.Println(err)
				response = blockResultsResponse{}
				response.Result.Height = strconv.FormatInt(block, 10)
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
