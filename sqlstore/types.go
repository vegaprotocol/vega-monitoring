package sqlstore

import (
	"errors"
	"fmt"
	"time"
)

const DefaultUpsertTxTimeout = 30 * time.Second

type StoreType string

const (
	StoreAssetPool             StoreType = "asset pool"
	StoreBlockSigner           StoreType = "block signer"
	StoreCometTxs              StoreType = "comet txs"
	StoreNetworkBalances       StoreType = "network balances"
	StoreNetworkHistorySegment StoreType = "network history segment"
	StoreMonitoringStatus      StoreType = "monitoring status"
)

var (
	ErrAcquireTx    = errors.New("failed to acquire transaction")
	ErrUpsertSingle = errors.New("failed to upsert single row")
	ErrUpsertCommit = errors.New("failed to commit the upsert transaction")
)

func NewUpsertErr(store StoreType, baseError, actualError error) error {
	return errors.Join(fmt.Errorf("failed to upsert %s: %w", store, baseError), actualError)
}
