package healthcheck

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus"
	"github.com/vegaprotocol/vega-monitoring/prometheus/nodescanner"
	"go.uber.org/zap"
)

type StartArgs struct {
	*HealthcheckArgs
	ExtendedMetrics bool
}

var startArgs StartArgs

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start healthcheck",
	Long:  `Start healthcheck`,
	Run: func(cmd *cobra.Command, args []string) {
		run(startArgs)
	},
}

func init() {
	HealthcheckCmd.AddCommand(startCmd)
	startArgs.HealthcheckArgs = &healthcheckArgs
	startCmd.PersistentFlags().BoolVar(&startArgs.ExtendedMetrics, "extended-metrics", false, "Collect extended metrics from non local node")
}

func run(args StartArgs) {
	//
	// SETUP
	//
	ctx, cancel := context.WithCancel(context.Background())
	var shutdown_wg sync.WaitGroup

	config, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		log.Fatalf("Failed to setup Services %+v\n", err)
	}

	prometheusService := prometheus.NewPrometheusService(&config.Prometheus)

	nodeScannerService := nodescanner.NewNodeScannerService(
		&config.Monitoring, prometheusService.VegaMonitoringCollector, logger,
	)

	//
	// start: Prometheus Endpoint
	//
	if config.Prometheus.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			logger.Info("Starting Prometheus Endpoint service")
			if err := prometheusService.Start(); err != nil {
				logger.Error("Failed to start Prometheus Endpoint", zap.Error(err))
				cancel()
			}
		}()
	} else {
		logger.Info("Not starting Prometheus Endpoint", zap.String("config", "Enabled=false"))
	}
	//
	// start: Node Scanner
	//
	if config.Prometheus.Enabled && args.ExtendedMetrics { // same flag as for Prometheus
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			logger.Info("Starting Node Scanner service in 10sec")
			time.Sleep(10 * time.Second)
			if err := nodeScannerService.Start(ctx); err != nil {
				logger.Error("Failed to start Node Scanner service", zap.Error(err))
				cancel()
			}
		}()
	} else {
		logger.Info("Not starting Node Scanner service", zap.String("config", "Enabled=false"))
	}

	logger.Info("Service has started")

	//
	// wait: For SIGNALL
	//
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigc
	logger.Info("Signal received, shutting down", zap.Any("signal", s))

	//
	// Send CANCEL to all services
	//
	cancel()

	//
	// shutdown: Prometheus Endpoint
	//
	if config.Prometheus.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			logger.Info("Shutting down Prometheus Endpoint")
			if err := prometheusService.Shutdown(ctx); err != nil {
				logger.Error("failed to stop Promethus Endpoint", zap.Error(err))
			}
		}()
	}

	//
	// Notify when all go routines are done
	//
	waitCh := make(chan struct{})
	go func() {
		defer close(waitCh)
		shutdown_wg.Wait()
		waitCh <- struct{}{}
	}()

	//
	// wait: For services to stop OR timeout
	//
	select {
	case <-waitCh:
		logger.Info("Evertything closed nicely\n")
	case <-time.After(5 * time.Second):
		logger.Error("Service timed out to stop. Force stopping\n")
	}

	//
	// DONE
	//
	time.Sleep(time.Millisecond * 100)
	logger.Info("Service has stopped")
}
