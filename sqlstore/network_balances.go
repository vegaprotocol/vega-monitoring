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

var ignoredWithdrawalIDs = []string{
	// Exploit for LDO market
	"cf3d77a9e5767132a4da41d534c28ae0f9749372cfb1b902cc3439d1649d1d33",
	"664cece6d582e534f818002cd5d9fc84df9f66f040bcfec92929fd176ed8a6ec",
	"3e1c058594bdd27a7b7b4670305c5b11480cb223e7a791b015e0411a5530564d",
}

var ignoredDepositsIDs = []string{
	// Duplicated deposits
	"8808d9ddd6c09593a519f4ad1c7117c247783a22c57cf6d73448ff9094552e07", // USDT: 12000000000
	"fdf81953b1c46b7bcac177d95ab849025a85bc3fbb664c5bcde64ccb7a63f24f", // USDC: 1000000000
	"f8546c2e6b59d78603155d8f857b4ceefccb7330743398d677d7eb49c9ee08e5", // USDT: 186000000
	"c759b728bd6b136c775dfc4eaf97ecb1ab78f0b08c91e76bef522004e6494dbd", // USDT: 87000000
	"008b0e9738bd3eb84271149ba34426cc7e92f1649e4a4af689aa23f3362d37f6", // USDT: 462893962
	"b461fed538c64a12a45e25b08b6da8e7f9edbe65af6d7418b4e97b4ebab8f753", // USDT: 12000000000
}

func NewNetworkBalances(connectionSource *vega_sqlstore.ConnectionSource) *NetworkBalances {
	return &NetworkBalances{
		ConnectionSource: connectionSource,
	}
}

func (nhs *NetworkBalances) Add(newBalance entities.NetworkBalance) {
	nhs.NetworkBalances = append(nhs.NetworkBalances, newBalance)
}

func (nhs *NetworkBalances) Upsert(ctx context.Context, newBalance entities.NetworkBalance) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
		    chain_id,
			balance_source,
			balance)
		VALUES ( $1, $2, $3, $4, $5 )
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
		newBalance.BalanceTime,
		newBalance.AssetID,
		newBalance.ChainID,
		newBalance.BalanceSource,
		newBalance.Balance,
	)

	return err
}

func (c *NetworkBalances) FlushUpsert(ctx context.Context) ([]entities.NetworkBalance, error) {
	blockCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		c.NetworkBalances = nil
	}()

	blockCtx, err := c.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreNetworkBalances, ErrAcquireTx, err)
	}

	for _, tx := range c.NetworkBalances {
		if err := c.Upsert(blockCtx, tx); err != nil {
			return nil, NewUpsertErr(StoreNetworkBalances, ErrUpsertSingle, err)
		}
	}

	if err := c.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreNetworkBalances, ErrUpsertCommit, err)
	}

	flushed := c.NetworkBalances

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
			chain_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), a.id, a.chain_id, 'PARTIES_TOTAL', COALESCE(b.balance, 0)
			FROM assets_current a
			LEFT JOIN latest_balance b ON (b.asset_id = a.id)
			GROUP BY a.id, a.chain_id, b.balance
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}

func (nhs *NetworkBalances) UpsertUnrealisedWithdrawalsBalance(ctx context.Context) error {
	ignoredIDs := PrepareListForInCondition(ignoredWithdrawalIDs)
	sqlQuery := `
	INSERT INTO metrics.network_balances (
		balance_time,
		asset_id,
	    chain_id,
		balance_source,
		balance)
	SELECT DATE_TRUNC('minute', NOW()), a.id, a.chain_id, 'UNREALISED_WITHDRAWALS_TOTAL', COALESCE(SUM(w.amount), 0)
		FROM assets_current a
		LEFT JOIN withdrawals_current w ON (
				w.asset = a.id
			AND w.withdrawn_timestamp = '1970-01-01 00:00:00'::timestamptz
			AND w.status = 'STATUS_FINALIZED'
			AND encode(w.id::bytea, 'hex') NOT IN (` + ignoredIDs + `)
		)
		GROUP BY a.id, a.chain_id
	ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
	SET
		balance=EXCLUDED.balance`

	_, err := nhs.Connection.Exec(ctx, sqlQuery)

	return err
}

func (nhs *NetworkBalances) UpsertUnfinalizedDeposits(ctx context.Context) error {
	ignoredIDs := PrepareListForInCondition(ignoredDepositsIDs)
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_balances (
			balance_time,
			asset_id,
		    chain_id,
			balance_source,
			balance)
		SELECT DATE_TRUNC('minute', NOW()), a.id, a.chain_id, 'UNFINALIZED_DEPOSITS', COALESCE(SUM(d.amount), 0)
			FROM assets_current a
			LEFT JOIN deposits_current d ON (d.asset = a.id AND d.status <> 'STATUS_FINALIZED') 
				AND encode(d.id::bytea, 'hex') NOT IN (`+ignoredIDs+`)
			GROUP BY a.id, a.chain_id
		ON CONFLICT (balance_time, asset_id, balance_source) DO UPDATE
		SET
			balance=EXCLUDED.balance`,
	)

	return err
}
