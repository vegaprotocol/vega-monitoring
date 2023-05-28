package sqlstore

import (
	"context"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/jackc/pgx/v4"
	"github.com/vegaprotocol/data-metrics-store/entities"
)

type BlockSigner struct {
	*vega_sqlstore.ConnectionSource
	columns     []string
	table_name  string
	blockSigner []*entities.BlockSigner
}

func NewBlockSigner(connectionSource *vega_sqlstore.ConnectionSource) *BlockSigner {
	return &BlockSigner{
		ConnectionSource: connectionSource,
		columns: []string{
			"vega_time", "height", "role", "tendermint_pub_key",
		},
		table_name: "m_block_signers",
	}
}

func (bs *BlockSigner) Add(data *entities.BlockSigner) {
	bs.blockSigner = append(bs.blockSigner, data)
}

func (bs *BlockSigner) Flush(ctx context.Context) ([]*entities.BlockSigner, error) {
	rows := make([][]interface{}, 0, len(bs.blockSigner))
	for _, data := range bs.blockSigner {
		rows = append(rows, []interface{}{
			data.VegaTime, data.Height, data.Role, data.TmPubKey,
		})
	}
	if rows != nil {
		copyCount, err := bs.Connection.CopyFrom(
			ctx,
			pgx.Identifier{bs.table_name}, bs.columns, pgx.CopyFromRows(rows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to copy block signer into database:%w", err)
		}

		if copyCount != int64(len(rows)) {
			return nil, fmt.Errorf("copied %d block signer rows into the database, expected to copy %d", copyCount, len(rows))
		}
	}

	flushed := bs.blockSigner
	bs.blockSigner = nil

	return flushed, nil
}

func (bs *BlockSigner) Upsert(ctx context.Context, newBlockSigner *entities.BlockSigner) error {
	_, err := bs.Connection.Exec(ctx, `
		INSERT INTO m_block_signers (
			vega_time,
			height,
			role,
			tendermint_address,
			tendermint_pub_key)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT (vega_time, role, tendermint_address) DO UPDATE
		SET
			height = EXCLUDED.height,
			tendermint_pub_key = EXCLUDED.tendermint_pub_key`,
		newBlockSigner.VegaTime,
		newBlockSigner.Height,
		newBlockSigner.Role,
		newBlockSigner.TmAddress,
		newBlockSigner.TmPubKey,
	)

	return err
}

func (bs *BlockSigner) FlushUpsert(ctx context.Context) ([]*entities.BlockSigner, error) {
	var blockCtx context.Context
	var cancel context.CancelFunc
	blockCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	blockCtx, err := bs.WithTransaction(blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to add transaction to context:%w", err)
	}

	for _, data := range bs.blockSigner {
		if err := bs.Upsert(blockCtx, data); err != nil {
			return nil, err
		}
	}

	if err := bs.Commit(blockCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction to FlushUpsert Block Signers: %w", err)
	}

	flushed := bs.blockSigner
	bs.blockSigner = nil

	return flushed, nil
}
