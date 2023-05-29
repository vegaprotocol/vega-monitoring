package read

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
)

type ReadService struct {
	cometClient      *comet.CometClient
	storeReadService StoreReadService
	log              *logging.Logger
}

type StoreReadService interface {
}

func NewReadService(
	cometClient *comet.CometClient,
	storeReadService StoreReadService,
	log *logging.Logger,
) (*ReadService, error) {
	return &ReadService{
		cometClient:      cometClient,
		storeReadService: storeReadService,
		log:              log,
	}, nil
}
