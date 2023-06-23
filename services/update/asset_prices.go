package update

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (us *UpdateService) UpdateAssetPrices(ctx context.Context) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdateAssetPrices"))

	logger.Info("Update Asset Prices: start")

	logger.Debug("reading asset price")
	prices, err := us.readService.GetAssetPrices()
	if err != nil {
		return err
	}

	logger.Debugf("found %d prices", len(prices))
	assetPricesStore := us.storeService.NewAssetPrices()
	for i := range prices {
		assetPricesStore.Add(&prices[i])
	}

	logger.Debug("flushing asset prices")
	storedPrices, err := assetPricesStore.FlushUpsert(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Prices, %w", err)
	}
	logger.Info(
		"Stored Asset Prices in SQLStore",
		zap.Int("row count", len(storedPrices)),
	)

	return nil
}
