package cmd

import (
	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/clients/coingecko"
	"github.com/vegaprotocol/vega-monitoring/clients/comet"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	vegaclient "github.com/vegaprotocol/vega-monitoring/clients/vega"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/metamonitoring"
	"github.com/vegaprotocol/vega-monitoring/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/ethereummonitoring"
	"github.com/vegaprotocol/vega-monitoring/prometheus/ethnodescanner"
	metamonitoringprom "github.com/vegaprotocol/vega-monitoring/prometheus/metamonitoring"
	"github.com/vegaprotocol/vega-monitoring/prometheus/nodescanner"
	"github.com/vegaprotocol/vega-monitoring/services"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"github.com/vegaprotocol/vega-monitoring/services/update"
)

type AllServices struct {
	Config                      *config.Config
	Log                         *logging.Logger
	StoreService                *services.StoreService
	ReadService                 *read.ReadService
	UpdateService               *update.UpdateService
	PrometheusService           *prometheus.PrometheusService
	NodeScannerService          *nodescanner.NodeScannerService
	EthereumMonitoringService   *ethereummonitoring.EthereumMonitoringService
	MetaMonitoringStatusService *metamonitoringprom.MetaMonitoringStatusService
	EthereumNodeScannerService  *ethnodescanner.EthNodeScannerService
	MonitoringService           metamonitoring.MetamonitoringService
}

func SetupServices(configFilePath string, forceDebug bool) (svc AllServices, err error) {
	svc.Config, svc.Log, err = config.GetConfigAndLogger(configFilePath, forceDebug)
	if err != nil {
		return
	}
	coingeckoClient := coingecko.NewCoingeckoClient(&svc.Config.Coingecko, svc.Log)
	cometClient := comet.NewCometClient(&svc.Config.CometBFT)
	coreClient := vegaclient.NewVegaClient(svc.Config.VegaCore.ApiURL)

	if svc.Config.DataNodeDBExtension.Enabled {
		var ethClient *ethutils.EthClient
		ethClient, err = ethutils.NewEthClient(svc.Config.Ethereum.RPCEndpoint, svc.Log)
		if err != nil {
			return
		}

		svc.StoreService, err = services.NewStoreService(&svc.Config.SQLStore, svc.Log)
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

		svc.MonitoringService, err = metamonitoring.NewMonitoringStatusUpdateService(svc.StoreService, coreClient, svc.Log)
		if err != nil {
			return
		}
	} else {
		svc.MonitoringService = metamonitoring.NewNopService()
	}

	if svc.Config.Prometheus.Enabled {
		svc.PrometheusService = prometheus.NewPrometheusService(&svc.Config.Prometheus)

		svc.NodeScannerService = nodescanner.NewNodeScannerService(
			&svc.Config.Monitoring, svc.PrometheusService.VegaMonitoringCollector, svc.Log,
		)

		svc.EthereumNodeScannerService = ethnodescanner.NewEthNodeScannerService(
			svc.Config.Monitoring.EthereumNode, svc.PrometheusService.VegaMonitoringCollector, svc.Log,
		)

		svc.EthereumMonitoringService = ethereummonitoring.NewEthereumMonitoringService(
			svc.Config.Monitoring.EthereumChain, svc.PrometheusService.VegaMonitoringCollector, svc.Log,
		)

		if svc.Config.DataNodeDBExtension.Enabled {
			svc.MetaMonitoringStatusService = metamonitoringprom.NewMetaMonitoringStatusService(
				svc.ReadService, svc.PrometheusService.VegaMonitoringCollector, svc.Log,
			)
		}
	}
	return
}
