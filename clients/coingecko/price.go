package coingecko

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

const timeout = 5 * time.Second

type PriceData struct {
	AssetSymbol string
	PriceUSD    decimal.Decimal
	Time        time.Time
}

func (c *CoingeckoClient) roundRobinApiKey() string {
	if len(c.config.ApiKeys) < 1 {
		return NoApiKey
	}

	n := int(c.idx.Load()) % len(c.config.ApiKeys)
	c.idx.Add(1)

	return c.config.ApiKeys[n]
}

func (c *CoingeckoClient) GetAssetPrices(assetIds []string) ([]PriceData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	apiKey := c.roundRobinApiKey()
	response, err := c.requestSimplePrice(ctx, assetIds, apiKey)
	if err != nil {
		// Retry without API Key if We initially tried with API key
		if apiKey != NoApiKey {
			var err2 error
			response, err2 = c.requestSimplePrice(ctx, assetIds, NoApiKey)
			if err2 != nil {
				return nil, errors.Join(err2, err)
			}
		} else {
			// Return error, we did not provide API key, so retry makes no sense
			return nil, err
		}
	}
	result := []PriceData{}
	for assetSymbol, data := range response {
		result = append(result, PriceData{
			AssetSymbol: assetSymbol,
			PriceUSD:    data.PriceUSD,
			Time:        time.Unix(data.LastUpdatedAt, 0),
		})
	}
	return result, nil
}

func (c *CoingeckoClient) GetPrices() ([]PriceData, error) {
	ids := []string{}
	coingeckoIdToVegaSymbol := map[string]string{}

	for vegaSymbol, coingeckoId := range c.config.AssetIds {
		ids = append(ids, coingeckoId)
		coingeckoIdToVegaSymbol[coingeckoId] = vegaSymbol
	}

	result, err := c.GetAssetPrices(ids)
	if err != nil {
		return nil, err
	}

	for i, assetPrice := range result {
		result[i].AssetSymbol = coingeckoIdToVegaSymbol[assetPrice.AssetSymbol]
	}

	return result, nil
}
