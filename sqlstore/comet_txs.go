package sqlstore

import (
	"context"
	"errors"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/vegaprotocol/vega-monitoring/clients/comet"
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
	blockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	blockCtx, err := c.WithTransaction(blockCtx)
	if err != nil {
		// We cannot keep those rows in memory because they will be added again
		// and at some point program hangs
		c.cometTxs = nil
		return nil, NewUpsertErr(StoreCometTxs, ErrAcquireTx, err)
	}

	for _, tx := range c.cometTxs {
		if err := c.UpsertWithoutTime(blockCtx, tx); err != nil {
			c.cometTxs = nil
			return nil, NewUpsertErr(StoreCometTxs, ErrUpsertSingle, err)
		}
	}

	if err := c.Commit(blockCtx); err != nil {
		c.cometTxs = nil
		return nil, NewUpsertErr(StoreCometTxs, ErrUpsertCommit, err)
	}

	flushed := c.cometTxs
	c.cometTxs = nil

	return flushed, nil
}

func (c *CometTxs) GetLatestBlockInStore(ctx context.Context) (int64, error) {
	result := &struct {
		Height *int64 `db:"height"`
	}{}

	if err := pgxscan.Get(ctx, c.Connection, result,
		`SELECT MAX(height) as height
		FROM metrics.comet_txs`,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	if result.Height == nil {
		return 0, nil
	}

	return *result.Height, nil
}
