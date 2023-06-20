package update

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/services"
	"github.com/vegaprotocol/vega-monitoring/services/read"
)

type UpdateService struct {
	readService  *read.ReadService
	storeService *services.StoreService
	log          *logging.Logger
}

func NewUpdateService(
	readService *read.ReadService,
	storeService *services.StoreService,
	log *logging.Logger,
) (*UpdateService, error) {
	return &UpdateService{
		readService:  readService,
		storeService: storeService,
		log:          log,
	}, nil
}
