package binance

import (
	"fmt"
	"math/big"
	"strings"
)

func (c *BinanceClient) GetAssetPrice(asset string) (*big.Float, error) {
	symbol := fmt.Sprintf("%sUSDT", strings.ToUpper(asset))
	response, err := c.requestAvgPrice(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get Price for asset %s, %w", asset, err)
	}
	balance, ok := new(big.Float).SetString(response.Price)
	if !ok {
		return nil, fmt.Errorf("failed to get Price for asset %s, failed to parse price %s, %w", asset, response.Price, err)
	}
	return balance, nil
}
