package update

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	dnentities "code.vegaprotocol.io/vega/datanode/entities"

	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/entities"
)

func (us *UpdateService) UpdateAssetPoolBalances(ctx context.Context, ethConfig, arbitrumConfig config.EthereumConfig) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdateAssetPoolBalances"))

	logger.Debug("Update Asset Pool Balances: start")

	// TODO: :et's think about passing clients to this function because now they are created over and over again.
	ethClient, err := ethutils.NewEthClient(ethConfig.RPCEndpoint, logger.Named("ethereum"))
	if err != nil {
		return err
	}

	arbitrumClient, err := ethutils.NewEthClient(arbitrumConfig.RPCEndpoint, logger.Named("arbitrum"))
	if err != nil {
		return err
	}

	ethChainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve Ethereum chain ID: %w", err)
	}

	arbitrumChainID, err := arbitrumClient.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve Arbitrum chain ID: %w", err)
	}

	assetsService := us.storeService.NewAssets()

	logger.Debug("Getting all assets for the network")
	assets, err := assetsService.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Pool Balances, failed to get assets from SQLStore: %w", err)
	}
	logger.Debugf("Got %d assets on the network", len(assets))

	now := time.Now().UTC().Truncate(time.Minute)
	networkBalancesStore := us.storeService.NewNetworkBalances()

	for _, asset := range assets {
		if asset.ERC20Contract == "" || asset.Status != dnentities.AssetStatusEnabled {
			continue
		}

		var fn func(assetContract string) (*big.Int, error)
		switch asset.ChainID {
		case arbitrumChainID:
			fn = func(assetContract string) (*big.Int, error) {
				return us.readService.GetAssetPoolBalanceForTokenFromArbitrum(assetContract, arbitrumConfig.AssetPoolAddress)
			}
		case ethChainID:
			fn = func(assetContract string) (*big.Int, error) {
				return us.readService.GetAssetPoolBalanceForTokenFromEthereum(assetContract, ethConfig.AssetPoolAddress)
			}
		default:
			return fmt.Errorf("asset's chain ID doesn't match any of chain IDs retrieved from blockchains: expecting one of [%v, %v], got %v", ethChainID, arbitrumChainID, asset.ChainID)
		}

		logger.Debug("Getting balance on the asset-pool", zap.String("asset", asset.Name))
		balance, err := fn(asset.ERC20Contract)
		if err != nil {
			return fmt.Errorf("failed to update Asset Pool Balances, failed to get balance for asset '%s' (%s): %w", asset.Name, asset.ERC20Contract, err)
		}

		logger.Debug("Got balance on the asset-pool", zap.String("asset", asset.Name), zap.String("balance", balance.String()))
		decimalBalance := decimal.NewFromBigInt(balance, 0)
		networkBalancesStore.Add(entities.NewAssetPoolBalance(asset.ID, now, asset.ERC20Contract, asset.ChainID, decimalBalance))
	}

	logger.Debug("Flushing balances to store")
	balances, err := networkBalancesStore.FlushUpsert(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Asset Pool Balances: %w", err)
	}
	logger.Debug(
		"Stored Asset Pool Balances in SQLStore",
		zap.Int("row count", len(balances)),
	)

	return nil
}

func (us *UpdateService) UpdatePartiesTotalBalances(ctx context.Context) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdatePartiesTotalBalances"))

	logger.Debug("Update Parties Total Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertPartiesTotalBalance(ctx); err != nil {
		logger.Error("Failed to update Parties Total Balances", zap.Error(err))
		return fmt.Errorf("failed to update Parties Total Balances: %w", err)
	}
	logger.Debug("Stored Parties Total Balances in SQLStore")
	return nil
}

func (us *UpdateService) UpdateUnrealisedWithdrawalsBalances(ctx context.Context) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdateUnrealisedWithdrawalsBalances"))

	logger.Debug("Update Unrealised Withdrawals Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertUnrealisedWithdrawalsBalance(ctx); err != nil {
		logger.Error("Failed to update Unrealised Withdrawals Balances", zap.Error(err))
		return fmt.Errorf("failed to update Unrealised Withdrawals Balances: %w", err)
	}
	logger.Debug("Stored Unrealised Withdrawals Balances in SQLStore")
	return nil
}

func (us *UpdateService) UpdateUnfinalizedDepositsBalances(ctx context.Context) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdateUnfinalizedDepositsBalances"))

	logger.Debug("Update Unfinalized Deposits Balances: start")

	networkBalancesStore := us.storeService.NewNetworkBalances()
	if err := networkBalancesStore.UpsertUnfinalizedDeposits(ctx); err != nil {
		logger.Error("Failed to update Unfinalized Deposits Balances", zap.Error(err))
		return fmt.Errorf("failed to update Unfinalized Deposits Balances: %w", err)
	}
	logger.Debug("Stored Unfinalized Deposits Balances in SQLStore")
	return nil
}
