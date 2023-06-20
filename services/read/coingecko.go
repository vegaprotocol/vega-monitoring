package read

import "github.com/vegaprotocol/vega-monitoring/clients/coingecko"

func (s *ReadService) GetAssetPrices() ([]coingecko.PriceData, error) {
	return s.coingeckoClient.GetPrices()
}
