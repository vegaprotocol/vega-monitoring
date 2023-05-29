package read

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
	"github.com/vegaprotocol/data-metrics-store/services"
)

type ReadService struct {
	cometClient  *comet.CometClient
	storeService *services.StoreService
	log          *logging.Logger
}

func NewReadService(
	cometClient *comet.CometClient,
	storeService *services.StoreService,
	log *logging.Logger,
) (*ReadService, error) {
	return &ReadService{
		cometClient:  cometClient,
		storeService: storeService,
		log:          log,
	}, nil
}
