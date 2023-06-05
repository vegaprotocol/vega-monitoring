package sqlstore

import (
	"context"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/vegaprotocol/data-metrics-store/entities"
)

type NetworkBalances struct {
	*vega_sqlstore.ConnectionSource
	NetworkBalances []entities.NetworkBalance
}

func NewNetworkBalances(connectionSource *vega_sqlstore.ConnectionSource) *NetworkBalances {
	return &NetworkBalances{
		ConnectionSource: connectionSource,
	}
}

func (nhs *NetworkBalances) Add(newBalance entities.NetworkBalance) {
	nhs.NetworkBalances = append(nhs.NetworkBalances, newBalance)
}

func (nhs *NetworkBalances) UpsertWithoutAssetId(ctx context.Context, newBalance entities.NetworkBalance) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		VALUES
			(
				$1,
				(SELECT id FROM assets_current WHERE erc20_contract = $2),
				$3,
				$4
			)
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
		newBalance.BalanceTime,
		newBalance.AssetEthereumHexAddress,
		newBalance.BalanceSource,
		newBalance.Balance,
	)

	return err
}

func (c *NetworkBalances) FlushUpsert(ctx context.Context) ([]entities.NetworkBalance, error) {
	var blockCtx context.Context
	var cancel context.CancelFunc
	blockCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	blockCtx, err := c.WithTransaction(blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to Flush Upsert Network Balances, failed to add transaction to context:%w", err)
	}

	for _, tx := range c.NetworkBalances {
		if err := c.UpsertWithoutAssetId(blockCtx, tx); err != nil {
			return nil, err
		}
	}

	if err := c.Commit(blockCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction to Flush Upsert Network Balances: %w", err)
	}

	flushed := c.NetworkBalances
	c.NetworkBalances = nil

	return flushed, nil
}

func (nhs *NetworkBalances) UpsertPartiesTotalBalance(ctx context.Context) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), accounts.asset_id, 'PARTIES_TOTAL', SUM(current_balances.balance)
			FROM current_balances, accounts
			WHERE current_balances.account_id = accounts.id
			GROUP BY accounts.asset_id
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}

func (nhs *NetworkBalances) UpsertUnrealisedWithdrawalsBalance(ctx context.Context) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), asset, 'UNREALISED_WITHDRAWALS_TOTAL', SUM(amount)
			FROM withdrawals
			WHERE
				withdrawn_timestamp = '1970-01-01 00:00:00'::timestamptz
			AND
				status = 'STATUS_FINALIZED'
			GROUP BY asset
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}

func (nhs *NetworkBalances) UpsertUnfinalizedDeposits(ctx context.Context) error {
	_, err := nhs.Connection.Exec(ctx, `
		WITH open_deposits AS (
			SELECT asset, SUM(amount) AS amount
			FROM deposits
			WHERE status = 'STATUS_OPEN'
			GROUP BY asset 
		), finalized_deposits AS (
			SELECT asset, SUM(amount) AS amount
			FROM deposits
			WHERE status = 'STATUS_FINALIZED'
			GROUP BY asset 
		)
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), open_deposits.asset, 'UNFINALIZED_DEPOSITS', open_deposits.amount - finalized_deposits.amount
		FROM open_deposits, finalized_deposits
		WHERE
			open_deposits.asset = finalized_deposits.asset
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}
