package update

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (us *UpdateService) UpdateAssetPrices(ctx context.Context) error {

	us.log.Info("Update Asset Prices: start")
	prices, err := us.readService.GetAssetPrices()
	if err != nil {
		return err
	}

	assetPricesStore := us.storeService.NewAssetPrices()
	for i, _ := range prices {
		assetPricesStore.Add(&prices[i])
	}

	storedPrices, err := assetPricesStore.FlushUpsert(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Prices, %w", err)
	}
	us.log.Info(
		"Stored Asset Prices in SQLStore",
		zap.Int("row count", len(storedPrices)),
	)

	return nil
}
