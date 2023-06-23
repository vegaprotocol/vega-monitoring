package update

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/services"
	"github.com/vegaprotocol/vega-monitoring/services/read"
)

const UpdaterType = "updater"

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
