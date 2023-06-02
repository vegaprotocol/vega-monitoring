package read

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
	"github.com/vegaprotocol/data-metrics-store/clients/ethutils"
)

type ReadService struct {
	cometClient      *comet.CometClient
	ethClient        *ethutils.EthClient
	storeReadService StoreReadService
	log              *logging.Logger
}

type StoreReadService interface {
}

func NewReadService(
	cometClient *comet.CometClient,
	ethClient *ethutils.EthClient,
	storeReadService StoreReadService,
	log *logging.Logger,
) (*ReadService, error) {
	return &ReadService{
		cometClient:      cometClient,
		ethClient:        ethClient,
		storeReadService: storeReadService,
		log:              log,
	}, nil
}
