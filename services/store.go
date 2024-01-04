package services

import (
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

type StoreService struct {
	config *config.SQLStoreConfig

	connSource *vega_sqlstore.ConnectionSource

	log *logging.Logger
}

func NewStoreService(
	config *config.SQLStoreConfig,
	log *logging.Logger,
) (*StoreService, error) {

	connConfig := config.GetConnectionConfig()

	connSource, err := vega_sqlstore.NewTransactionalConnectionSource(log, connConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Connection Source, %w", err)
	}

	return &StoreService{
		config:     config,
		connSource: connSource,
		log:        log,
	}, nil
}

func (s *StoreService) NewBlockSigner() *sqlstore.BlockSigner {
	return sqlstore.NewBlockSigner(s.connSource)
}

func (s *StoreService) NewNetworkHistorySegment() *sqlstore.NetworkHistorySegment {
	return sqlstore.NewNetworkHistorySegment(s.connSource)
}

func (s *StoreService) NewCometTxs() *sqlstore.CometTxs {
	return sqlstore.NewCometTxs(s.connSource)
}

func (s *StoreService) NewNetworkBalances() *sqlstore.NetworkBalances {
	return sqlstore.NewNetworkBalances(s.connSource)
}

func (s *StoreService) NewAssetPrices() *sqlstore.AssetPrices {
	return sqlstore.NewAssetPrices(s.connSource)
}

func (s *StoreService) NewMetamonitoringStatus() *sqlstore.MetamonitoringStatus {
	return sqlstore.NewMetamonitoringStatus(s.connSource)
}

func (s *StoreService) NewMonitoringStatus() *sqlstore.MonitoringStatus {
	return sqlstore.NewMonitoringStatus(s.connSource)
}

// Data Node tables
func (s *StoreService) NewAssets() *vega_sqlstore.Assets {
	return vega_sqlstore.NewAssets(s.connSource)
}
