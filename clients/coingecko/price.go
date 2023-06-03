package coingecko

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriceData struct {
	AssetSymbol string
	PriceUSD    decimal.Decimal
	Time        time.Time
}

func (c *CoingeckoClient) GetAssetPrices(assetIds []string) ([]PriceData, error) {
	response, err := c.requestSimplePrice(assetIds)
	if err != nil {
		return nil, err
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
