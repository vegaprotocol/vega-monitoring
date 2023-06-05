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

	us.log.Info("Update Asset Pool Balances: start")

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

func (us *UpdateService) UpdatePartiesTotalBalances(ctx context.Context) error {

	us.log.Info("Update Parties Total Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertPartiesTotalBalance(ctx); err != nil {
		us.log.Error("Failed to update Parties Total Balances", zap.Error(err))
		return fmt.Errorf("failed to update Parties Total Balances, %w", err)
	}
	us.log.Info("Stored Parties Total Balances in SQLStore")
	return nil
}

func (us *UpdateService) UpdateUnrealisedWithdrawalsBalances(ctx context.Context) error {

	us.log.Info("Update Unrealised Withdrawals Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertUnrealisedWithdrawalsBalance(ctx); err != nil {
		us.log.Error("Failed to update Unrealised Withdrawals Balances", zap.Error(err))
		return fmt.Errorf("failed to update Unrealised Withdrawals Balances, %w", err)
	}
	us.log.Info("Stored Unrealised Withdrawals Balances in SQLStore")
	return nil
}

func (us *UpdateService) UpdateUnfinalizedDepositsBalances(ctx context.Context) error {

	us.log.Info("Update Unfinalized Deposits Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertUnfinalizedDeposits(ctx); err != nil {
		us.log.Error("Failed to update Unfinalized Deposits Balances", zap.Error(err))
		return fmt.Errorf("failed to update Unfinalized Deposits Balances, %w", err)
	}
	us.log.Info("Stored Unfinalized Deposits Balances in SQLStore")
	return nil
}
