package sqlstore

import (
	"context"
	"fmt"
	"strings"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"

	"github.com/vegaprotocol/vega-monitoring/clients/coingecko"
)

type AssetPrices struct {
	*vega_sqlstore.ConnectionSource
	assetPrices []*coingecko.PriceData
}

func NewAssetPrices(connectionSource *vega_sqlstore.ConnectionSource) *AssetPrices {
	return &AssetPrices{
		ConnectionSource: connectionSource,
	}
}

func (ap *AssetPrices) Add(data *coingecko.PriceData) {
	ap.assetPrices = append(ap.assetPrices, data)
}

func (ap *AssetPrices) Upsert(ctx context.Context, newAssetPrices *coingecko.PriceData) error {
	_, err := ap.Connection.Exec(ctx, `
		INSERT INTO metrics.asset_prices (
			price_time,
		    price,
			asset_id
			)
		SELECT price_time, price, asset_id
		    FROM (
				SELECT $1::timestamptz as price_time, $2::numeric(16,8) as price, id as asset_id FROM assets_current WHERE LOWER(symbol) = $3
			) as assets_prices
		ON CONFLICT (price_time, asset_id) DO UPDATE
		SET
			price=EXCLUDED.price`,
		newAssetPrices.Time,
		newAssetPrices.PriceUSD,
		strings.ToLower(newAssetPrices.AssetSymbol),
	)

	if err != nil {
		return fmt.Errorf("could not get asset prices for %q: %w", newAssetPrices.AssetSymbol, err)
	}

	return nil
}

func (ap *AssetPrices) FlushUpsert(ctx context.Context) ([]*coingecko.PriceData, error) {
	blockCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		// We cannot keep those rows in memory because they will be added again
		// and at some point program hangs
		ap.assetPrices = nil
	}()

	blockCtx, err := ap.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreAssetPool, ErrAcquireTx, err)
	}

	for _, data := range ap.assetPrices {
		if err := ap.Upsert(blockCtx, data); err != nil {
			return nil, NewUpsertErr(StoreAssetPool, ErrUpsertSingle, err)
		}
	}

	if err := ap.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreAssetPool, ErrUpsertCommit, err)
	}

	flushed := ap.assetPrices

	return flushed, nil
}
