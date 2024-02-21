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

	latestSegmentsCache map[string]int64 // map[data-node-url]block-height // TODO: Make this struct or something...
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
