package coingecko

import (
	"time"
)

type PriceData struct {
	AssetId  string
	PriceUSD float64
	Time     time.Time
}

func (c *CoingeckoClient) GetAssetPrices(assetIds []string) ([]PriceData, error) {
	response, err := c.requestSimplePrice(assetIds)
	if err != nil {
		return nil, err
	}
	var result []PriceData
	for assetId, data := range response {
		result = append(result, PriceData{
			AssetId:  assetId,
			PriceUSD: data.PriceUSD,
			Time:     time.Unix(data.LastUpdatedAt, 0),
		})
	}
	return result, nil
}
