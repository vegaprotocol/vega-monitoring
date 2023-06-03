package read

import "github.com/vegaprotocol/data-metrics-store/clients/coingecko"

func (s *ReadService) GetAssetPrices() ([]coingecko.PriceData, error) {
	return s.coingeckoClient.GetPrices()
}
