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

func (b *Blocks) GetLatestBlockWithCache(ctx context.Context, cacheTime time.Duration) (*vega_entities.Block, error) {
	b.lastBlockMutex.Lock()
	defer b.lastBlockMutex.Unlock()

	now := time.Now()
	if b.lastBlockCache != nil && b.lastBlockCache.cachedAt.Sub(now) <= cacheTime {
		return b.lastBlockCache.block, nil
	}

	block := &vega_entities.Block{}
	if err := pgxscan.Get(ctx, b.Connection, block,
		`SELECT vega_time, height, hash
		FROM last_block`); err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	b.lastBlockCache.block = block
	b.lastBlockCache.cachedAt = now

	return block, nil

}
