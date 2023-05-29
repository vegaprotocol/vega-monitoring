package update

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/data-metrics-store/services"
	"github.com/vegaprotocol/data-metrics-store/services/read"
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
