package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	NoApiKey         = ""
	ApiKeyHeaderName = "x-cg-demo-api-key"
)

type coingeckoSimplePriceResponse map[string]struct {
	PriceUSD      decimal.Decimal `json:"usd"`
	LastUpdatedAt int64           `json:"last_updated_at"`
}

func (c *CoingeckoClient) requestSimplePrice(ctx context.Context, ids []string, apiKey string) (coingeckoSimplePriceResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed rate limiter to get Simple Price for ids: %s. %w", ids, err)
	}

	url := fmt.Sprintf(SimplePriceURL, c.config.ApiURL, strings.Join(ids, ","))
	//https://api.coingecko.com/api/v3/simple/price?ids=vega-protocol,tether,usd-coin,weth&vs_currencies=usd&include_last_updated_at=true
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed to create new response with context: %w", err)
	}

	if apiKey != NoApiKey {
		req.Header.Add(ApiKeyHeaderName, apiKey)
	}

	c.log.Debug("Sending Coingecko request", zap.String("url", url))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed to get simple price for ids(%s): %w", ids, err)
	}

	if resp.StatusCode != http.StatusOK {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("invalid response status code: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
	var payload coingeckoSimplePriceResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed to parse response for get simple price for ids(%s): %w", ids, err)
	}

	return payload, nil
}
