package read

import (
	"code.vegaprotocol.io/vega/logging"

	"github.com/vegaprotocol/vega-monitoring/clients/coingecko"
	"github.com/vegaprotocol/vega-monitoring/clients/comet"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

type ReadService struct {
	coingeckoClient  *coingecko.CoingeckoClient
	cometClient      *comet.CometClient
	storeReadService StoreReadService
	log              *logging.Logger
	ethClient        *ethutils.EthClient
	arbitrumClient   *ethutils.EthClient
}

type StoreReadService interface {
	NewNetworkHistorySegment() *sqlstore.NetworkHistorySegment
	NewMonitoringStatus() *sqlstore.MonitoringStatus
}

func NewReadService(coingeckoClient *coingecko.CoingeckoClient, cometClient *comet.CometClient, ethClient *ethutils.EthClient, arbitrumClient *ethutils.EthClient, storeReadService StoreReadService, log *logging.Logger) (*ReadService, error) {
	return &ReadService{
		coingeckoClient:  coingeckoClient,
		cometClient:      cometClient,
		ethClient:        ethClient,
		arbitrumClient:   arbitrumClient,
		storeReadService: storeReadService,
		log:              log,
	}, nil
}
