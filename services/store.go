package services

import (
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/data-metrics-store/config"
	"github.com/vegaprotocol/data-metrics-store/sqlstore"
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

	connConfig := vega_sqlstore.NewDefaultConfig().ConnectionConfig
	connConfig.Host = config.Host
	connConfig.Port = config.Port
	connConfig.Username = config.Username
	connConfig.Password = config.Password
	connConfig.Database = config.Database

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
