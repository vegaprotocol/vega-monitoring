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

type coingeckoSimplePriceResponse map[string]struct {
	PriceUSD      decimal.Decimal `json:"usd"`
	LastUpdatedAt int64           `json:"last_updated_at"`
}

func (c *CoingeckoClient) requestSimplePrice(ids []string) (coingeckoSimplePriceResponse, error) {
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed rate limiter to get Simple Price for ids: %s. %w", ids, err)
	}
	//https://api.coingecko.com/api/v3/simple/price?ids=vega-protocol,tether,usd-coin,weth&vs_currencies=usd&include_last_updated_at=true
	url := fmt.Sprintf("%s/simple/price?vs_currencies=usd&include_last_updated_at=true&ids=%s", c.config.ApiURL, strings.Join(ids, ","))
	c.log.Debug("Sending Coingecko request", zap.String("url", url))
	resp, err := http.Get(url)
	if err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed to get Simple Price for ids: %s. %w", ids, err)
	}
	defer resp.Body.Close()
	var payload coingeckoSimplePriceResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return coingeckoSimplePriceResponse{}, fmt.Errorf("failed to parse response for get Simple Price for ids: %s. %w", ids, err)
	}

	return payload, nil
}
