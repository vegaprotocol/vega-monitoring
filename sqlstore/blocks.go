package sqlstore

import (
	"context"
	"fmt"
	"sync"
	"time"

	vega_entities "code.vegaprotocol.io/vega/datanode/entities"
	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"
)

type cachedBlock struct {
	cachedAt time.Time
	block    *vega_entities.Block
}

type Blocks struct {
	*vega_sqlstore.ConnectionSource

	lastBlockMutex sync.Mutex
	lastBlockCache *cachedBlock
}

func NewBlocks(connectionSource *vega_sqlstore.ConnectionSource) *Blocks {
	return &Blocks{
		ConnectionSource: connectionSource,
	}
}

func (b *Blocks) GetLastBlock(ctx context.Context) (*vega_entities.Block, error) {
	block := &vega_entities.Block{}
	if err := pgxscan.Get(ctx, b.Connection, block,
		`SELECT vega_time, height, hash
		FROM last_block`); err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return block, nil
}

func (b *Blocks) GetLatestBlockWithCache(ctx context.Context, cacheTime time.Duration) (*vega_entities.Block, error) {
	b.lastBlockMutex.Lock()
	defer b.lastBlockMutex.Unlock()

	now := time.Now()
	if b.lastBlockCache != nil && now.Sub(b.lastBlockCache.cachedAt) <= cacheTime {
		return b.lastBlockCache.block, nil
	}

	block := &vega_entities.Block{}
	if err := pgxscan.Get(ctx, b.Connection, block,
		`SELECT vega_time, height, hash
		FROM last_block`); err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	if b.lastBlockCache == nil {
		b.lastBlockCache = &cachedBlock{}
	}
	b.lastBlockCache.block = block
	b.lastBlockCache.cachedAt = now

	return block, nil

}
