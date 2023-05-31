package sqlstore

import (
	"context"
	"errors"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
)

type CometTxs struct {
	*vega_sqlstore.ConnectionSource
	cometTxs []comet.CometTx
}

func NewCometTxs(connectionSource *vega_sqlstore.ConnectionSource) *CometTxs {
	return &CometTxs{
		ConnectionSource: connectionSource,
	}
}

func (nhs *CometTxs) AddWithoutTime(newTx comet.CometTx) {
	nhs.cometTxs = append(nhs.cometTxs, newTx)
}

func (nhs *CometTxs) UpsertWithoutTime(ctx context.Context, newTx comet.CometTx) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.comet_txs (
			vega_time,
			height,
			height_idx,
			code,
			submitter,
			command,
			attributes,
			info)
		VALUES
			(
				(SELECT vega_time FROM blocks WHERE height = $1),
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8
			)
		ON CONFLICT (vega_time, height_idx) DO UPDATE
		SET
			height=EXCLUDED.height,
			code=EXCLUDED.code,
			submitter=EXCLUDED.submitter,
			command=EXCLUDED.command,
			attributes=EXCLUDED.attributes,
			info=EXCLUDED.info`,
		newTx.Height,
		newTx.Height,
		newTx.HeightIdx,
		newTx.Code,
		newTx.Submitter,
		newTx.Command,
		newTx.Attributes,
		newTx.Info,
	)

	return err
}

func (c *CometTxs) FlushUpsertWithoutTime(ctx context.Context) ([]comet.CometTx, error) {
	var blockCtx context.Context
	var cancel context.CancelFunc
	blockCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	blockCtx, err := c.WithTransaction(blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to add transaction to context:%w", err)
	}

	for _, tx := range c.cometTxs {
		if err := c.UpsertWithoutTime(blockCtx, tx); err != nil {
			return nil, err
		}
	}

	if err := c.Commit(blockCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction to FlushUpsertWithoutTime Network History Segments: %w", err)
	}

	flushed := c.cometTxs
	c.cometTxs = nil

	return flushed, nil
}

func (c *CometTxs) GetLastestBlockInStore(ctx context.Context) (int64, error) {

	result := &struct {
		Height int64 `db:"height"`
	}{}

	if err := pgxscan.Get(ctx, c.Connection, result,
		`SELECT height
		FROM metrics.comet_txs
		ORDER BY metrics.comet_txs.vega_time DESC
		LIMIT 1`,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return result.Height, nil
}