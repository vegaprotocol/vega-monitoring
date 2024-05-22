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
			asset_id,
			price)
		VALUES
			(
				$1,
				(SELECT id FROM assets_current WHERE LOWER(symbol) = $2),
				$3
			)
		ON CONFLICT (price_time, asset_id) DO UPDATE
		SET
			price=EXCLUDED.price`,
		newAssetPrices.Time,
		strings.ToLower(newAssetPrices.AssetSymbol),
		newAssetPrices.PriceUSD,
	)

	return fmt.Errorf("could not get asset prices for %q: %w", newAssetPrices.AssetSymbol, err)
}

func (ap *AssetPrices) FlushUpsert(ctx context.Context) ([]*coingecko.PriceData, error) {
	blockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	ap.assetPrices = nil

	return flushed, nil
}
