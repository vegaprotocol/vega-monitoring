package update

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/vegaprotocol/data-metrics-store/entities"
	"go.uber.org/zap"
)

func (us *UpdateService) UpdateAssetPoolBalances(ctx context.Context) error {

	assetsService := us.storeService.NewAssets()

	assets, err := assetsService.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Pool Balances, failed to get assets from SQLStore, %w", err)
	}

	now := time.Now().UTC().Truncate(time.Minute)
	networkBalancesStore := us.storeService.NewNetworkBalances()

	for _, asset := range assets {
		balance, err := us.readService.GetAssetPoolBalanceForToken(asset.ERC20Contract)
		if err != nil {
			return fmt.Errorf("failed to update Asset Pool Balances, failed to get balance for asset '%s' (%s), %w", asset.Name, asset.ERC20Contract, err)
		}
		decimalBalance := decimal.NewFromBigInt(balance, 0)
		networkBalancesStore.Add(entities.NewAssetPoolBalance(now, asset.ERC20Contract, decimalBalance))
	}

	balances, err := networkBalancesStore.FlushUpsert(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Pool Balances, %w", err)
	}
	us.log.Info(
		"Stored Asset Pool Balances in SQLStore",
		zap.Int("row count", len(balances)),
	)

	return nil
}
