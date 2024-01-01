package sqlstore

import (
	"context"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/vegaprotocol/vega-monitoring/entities"
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
	blockCtx, cancel := context.WithTimeout(ctx, DefaultUpsertTxTimeout)
	defer cancel()

	blockCtx, err := c.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreNetworkBalances, ErrAcquireTx, err)
	}

	for _, tx := range c.NetworkBalances {
		if err := c.UpsertWithoutAssetId(blockCtx, tx); err != nil {
			return nil, NewUpsertErr(StoreNetworkBalances, ErrUpsertSingle, err)
		}
	}

	if err := c.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreNetworkBalances, ErrUpsertCommit, err)
	}

	flushed := c.NetworkBalances
	c.NetworkBalances = nil

	return flushed, nil
}

func (nhs *NetworkBalances) UpsertPartiesTotalBalance(ctx context.Context) error {
	_, err := nhs.Connection.Exec(ctx, `
		WITH latest_balance AS (
			SELECT accounts.asset_id, SUM(current_balances.balance) AS balance
			FROM current_balances, accounts
			WHERE current_balances.account_id = accounts.id
			GROUP BY accounts.asset_id
		)
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), a.id, 'PARTIES_TOTAL', COALESCE(b.balance, 0)
			FROM assets_current a
			LEFT JOIN latest_balance b ON (b.asset_id = a.id)
			GROUP BY a.id, b.balance
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
		SELECT DATE_TRUNC('minute', NOW()), a.id, 'UNREALISED_WITHDRAWALS_TOTAL', COALESCE(SUM(w.amount), 0)
			FROM assets_current a
			LEFT JOIN withdrawals_current w ON (w.asset = a.id AND w.withdrawn_timestamp = '1970-01-01 00:00:00'::timestamptz AND w.status = 'STATUS_FINALIZED')
			GROUP BY a.id
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}

func (nhs *NetworkBalances) UpsertUnfinalizedDeposits(ctx context.Context) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), a.id, 'UNFINALIZED_DEPOSITS', COALESCE(SUM(d.amount), 0)
			FROM assets_current a
			LEFT JOIN deposits_current d ON (d.asset = a.id AND d.status <> 'STATUS_FINALIZED')
			GROUP BY a.id
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}
