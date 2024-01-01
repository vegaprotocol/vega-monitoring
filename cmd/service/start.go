package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/cmd"
	"github.com/vegaprotocol/vega-monitoring/pprof"
	"go.uber.org/zap"
)

type StartArgs struct {
	*ServiceArgs
	EnablePprof bool
}

var startArgs StartArgs

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start service",
	Long:  `Start service`,
	Run: func(cmd *cobra.Command, args []string) {
		startService(startArgs)
	},
}

func init() {
	ServiceCmd.AddCommand(startCmd)
	startArgs.ServiceArgs = &serviceArgs
	startCmd.PersistentFlags().BoolVar(&startArgs.EnablePprof, "enable-pprof", true, "Enables pprof server on port :6161")
}

func startService(args StartArgs) {
	//
	// SETUP
	//
	ctx, cancel := context.WithCancel(context.Background())
	var shutdown_wg sync.WaitGroup

	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		log.Fatalf("Failed to setup Services %+v\n", err)
	}

	if svc.Config.DataNodeDBExtension.Enabled {
		svc.Log.Debug("Starting DataNode DB Extension services", zap.Bool("DataNodeDBExtension.Enabled", true))
		startDataNodeDBExtension(&svc, &shutdown_wg, ctx)
	} else {
		svc.Log.Info("Not starting DataNode DB Extension services", zap.Bool("DataNodeDBExtension.Enabled", false))
	}

	if args.EnablePprof {
		go func() {
			svc.Log.Debug("Starting pprof server on port 6161")
			if err := pprof.StartPprofServer(":6161"); err != nil {
				panic(fmt.Errorf("failed to start pprof server: %w", err))
			}
		}()
	}

	if svc.Config.Prometheus.Enabled {
		//
		// start: Prometheus Endpoint
		//
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Starting Prometheus Endpoint service", zap.Bool("Prometheus.Enabled", true))
			if err := svc.PrometheusService.Start(); err != nil {
				svc.Log.Error("Failed to start Prometheus Endpoint", zap.Error(err))
				cancel()
			}
		}()

		//
		// start: Node Scanner
		//
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Starting Node Scanner service in 10sec", zap.Bool("Prometheus.Enabled", true))
			time.Sleep(10 * time.Second)
			if err := svc.NodeScannerService.Start(ctx); err != nil {
				svc.Log.Error("Failed to start Node Scanner service", zap.Error(err))
				cancel()
			}
		}()

		//
		// start: Ethereum Node Scanner
		//
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Starting Ethereum Node Scanner service in 20sec", zap.Bool("Prometheus.Enabled", true))
			time.Sleep(20 * time.Second)
			if err := svc.EthereumNodeScannerService.Start(ctx); err != nil {
				svc.Log.Error("Failed to start Ethereum Node Scanner service", zap.Error(err))
				cancel()
			}
		}()

		if svc.Config.DataNodeDBExtension.Enabled {
			//
			// start: MetaMonitoring Statuses
			//
			shutdown_wg.Add(1)
			go func() {
				defer shutdown_wg.Done()
				svc.Log.Info("Starting MetaMonitoring Statuses service in 15sec", zap.Bool("Prometheus.Enabled", true), zap.Bool("DataNodeDBExtension.Enabled", true))
				time.Sleep(15 * time.Second)
				if err := svc.MetaMonitoringStatusService.Start(ctx); err != nil {
					svc.Log.Error("Failed to start MetaMonitoring Statuses service", zap.Error(err))
					cancel()
				}
			}()
		}

	} else {
		svc.Log.Info("Not starting Prometheus Endpoint", zap.Bool("Prometheus.Enabled", false))
		svc.Log.Info("Not starting Node Scanner service", zap.Bool("Prometheus.Enabled", false))
	}
	//
	// start: example service
	//
	// shutdown_wg.Add(1)
	// go func() {
	// 	defer shutdown_wg.Done()
	// 	fmt.Printf("Starting something #2\n")
	// }()

	svc.Log.Info("Service has started")

	//
	// wait: For SIGNALL
	//
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigc
	svc.Log.Info("Signal received, shutting down", zap.Any("signal", s))

	//
	// Send CANCEL to all services
	//
	cancel()

	//
	// shutdown: Prometheus Endpoint
	//
	if svc.Config.Prometheus.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Shutting down Prometheus Endpoint")
			if err := svc.PrometheusService.Shutdown(ctx); err != nil {
				svc.Log.Error("failed to stop Promethus Endpoint", zap.Error(err))
			}
		}()
	}
	//
	// shutdown: example service
	//
	// shutdown_wg.Add(1)
	// go func() {
	// 	defer shutdown_wg.Done()
	// 	fmt.Printf("Shutting down something #1\n")
	// }()

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
		svc.Log.Info("Evertything closed nicely\n")
	case <-time.After(5 * time.Second):
		svc.Log.Error("Service timed out to stop. Force stopping\n")
	}

	//
	// DONE
	//
	time.Sleep(time.Millisecond * 100)
	svc.Log.Info("Service has stopped")
}
