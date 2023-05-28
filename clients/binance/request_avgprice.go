package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type binanceAvgPriceResponse struct {
	Price string `json:"price"`
}

func (c *BinanceClient) requestAvgPrice(symbol string) (binanceAvgPriceResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return binanceAvgPriceResponse{}, fmt.Errorf("failed rate limiter to get AvgPrice for symbol: %s. %w", symbol, err)
	}
	url := fmt.Sprintf("%s/avgPrice?symbol=%s", c.apiURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return binanceAvgPriceResponse{}, fmt.Errorf("failed to get AvgPrice for symbol: %s. %w", symbol, err)
	}
	defer resp.Body.Close()
	var payload binanceAvgPriceResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return binanceAvgPriceResponse{}, fmt.Errorf("failed to parse response for get AvgPrice for symbol: %s. %w", symbol, err)
	}

	return payload, nil
}
