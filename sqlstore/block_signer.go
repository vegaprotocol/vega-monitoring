package sqlstore

import (
	"context"
	"errors"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/vegaprotocol/vega-monitoring/entities"
)

type BlockSigner struct {
	*vega_sqlstore.ConnectionSource
	blockSigner []*entities.BlockSigner
}

func NewBlockSigner(connectionSource *vega_sqlstore.ConnectionSource) *BlockSigner {
	return &BlockSigner{
		ConnectionSource: connectionSource,
	}
}

func (bs *BlockSigner) Add(data *entities.BlockSigner) {
	bs.blockSigner = append(bs.blockSigner, data)
}

func (bs *BlockSigner) Flush(ctx context.Context) ([]*entities.BlockSigner, error) {
	rows := make([][]interface{}, 0, len(bs.blockSigner))
	for _, data := range bs.blockSigner {
		rows = append(rows, []interface{}{
			data.VegaTime, data.Role, data.TmPubKey,
		})
	}
	if rows != nil {
		copyCount, err := bs.Connection.CopyFrom(
			ctx,
			pgx.Identifier{"metrics.block_signers"},
			[]string{"vega_time", "role", "tendermint_pub_key"},
			pgx.CopyFromRows(rows),
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
		INSERT INTO metrics.block_signers (
			vega_time,
			role,
			tendermint_pub_key)
		VALUES
			($1, $2, $3)
		ON CONFLICT (vega_time, role, tendermint_pub_key) DO NOTHING`,
		newBlockSigner.VegaTime,
		newBlockSigner.Role,
		newBlockSigner.TmPubKey,
	)

	return err
}

func (bs *BlockSigner) FlushUpsert(ctx context.Context) ([]*entities.BlockSigner, error) {
	blockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	blockCtx, err := bs.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreBlockSigner, ErrAcquireTx, err)
	}

	for _, data := range bs.blockSigner {
		if err := bs.Upsert(blockCtx, data); err != nil {
			return nil, NewUpsertErr(StoreBlockSigner, ErrUpsertSingle, err)
		}
	}

	if err := bs.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreBlockSigner, ErrUpsertCommit, err)
	}

	flushed := bs.blockSigner
	bs.blockSigner = nil

	return flushed, nil
}

func (bs *BlockSigner) GetLastestBlockInStore(ctx context.Context) (int64, error) {

	result := &struct {
		Height int64 `db:"height"`
	}{}

	if err := pgxscan.Get(ctx, bs.Connection, result,
		`SELECT blocks.height
		FROM metrics.block_signers, blocks
		WHERE
			metrics.block_signers.vega_time = blocks.vega_time
		ORDER BY metrics.block_signers.vega_time DESC
		LIMIT 1`,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return result.Height, nil
}

// SELECT s.i AS missing_height
// FROM generate_series(1,(SELECT MAX(height) FROM m_block_signers)) s(i)
// LEFT OUTER JOIN m_block_signers ON (m_block_signers.height = s.i)
// WHERE m_block_signers.height IS NULL;

//SELECT MAX(height) FROM m_block_signers;
