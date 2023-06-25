package cmd

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/clients/coingecko"
	"github.com/vegaprotocol/vega-monitoring/clients/comet"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/nodescanner"
	"github.com/vegaprotocol/vega-monitoring/services"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"github.com/vegaprotocol/vega-monitoring/services/update"
)

type AllServices struct {
	Config             *config.Config
	Log                *logging.Logger
	StoreService       *services.StoreService
	ReadService        *read.ReadService
	UpdateService      *update.UpdateService
	PrometheusService  *prometheus.PrometheusService
	NodeScannerService *nodescanner.NodeScannerService
}

func SetupServices(configFilePath string, forceDebug bool) (svc AllServices, err error) {
	svc.Config, svc.Log, err = config.GetConfigAndLogger(configFilePath, forceDebug)
	if err != nil {
		return
	}

	svc.StoreService, err = services.NewStoreService(&svc.Config.SQLStore, svc.Log)
	if err != nil {
		return
	}

	coingeckoClient := coingecko.NewCoingeckoClient(&svc.Config.Coingecko, svc.Log)
	cometClient := comet.NewCometClient(&svc.Config.CometBFT)
	ethClient, err := ethutils.NewEthClient(&svc.Config.Ethereum, svc.Log)
	if err != nil {
		return
	}
	svc.ReadService, err = read.NewReadService(coingeckoClient, cometClient, ethClient, svc.StoreService, svc.Log)
	if err != nil {
		return
	}

	svc.UpdateService, err = update.NewUpdateService(svc.ReadService, svc.StoreService, svc.Log)
	if err != nil {
		return
	}

	svc.PrometheusService = prometheus.NewPrometheusService(&svc.Config.Prometheus)

	svc.NodeScannerService = nodescanner.NewNodeScannerService(
		&svc.Config.Monitoring, svc.PrometheusService.VegaMonitoringCollector, svc.Log,
	)
	return
}
