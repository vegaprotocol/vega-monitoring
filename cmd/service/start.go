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
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
	"go.uber.org/zap"
)

type StartArgs struct {
	*ServiceArgs
}

var startArgs StartArgs

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start service",
	Long:  `Start service`,
	Run: func(cmd *cobra.Command, args []string) {
		run(startArgs)
	},
}

func init() {
	ServiceCmd.AddCommand(startCmd)
	startArgs.ServiceArgs = &serviceArgs
}

func run(args StartArgs) {
	//
	// SETUP
	//
	ctx, cancel := context.WithCancel(context.Background())
	var shutdown_wg sync.WaitGroup

	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		log.Fatalf("Failed to setup Services %+v\n", err)
	}
	if err := sqlstore.MigrateToLatestSchema(svc.Log, svc.Config.SQLStore.GetConnectionConfig()); err != nil {
		log.Fatalf("Failed to migrate database to latest version %+v\n", err)
	}

	//
	// start: Block Singers Service
	//
	if svc.Config.Services.BlockSigners.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runBlockSignersScraper(ctx, &svc)
		}()
	} else {
		svc.Log.Info("Not starting Block Signers Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Network History Segments Service
	//
	if svc.Config.Services.NetworkHistorySegments.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runNetworkHistorySegmentsScraper(ctx, &svc)
		}()
	} else {
		svc.Log.Info("Not starting Network History Segments Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Comet Txs Service
	//
	if svc.Config.Services.CometTxs.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runCometTxsScraper(ctx, &svc)
		}()
	} else {
		svc.Log.Info("Not starting Comet Txs Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Network Balances
	//
	if svc.Config.Services.NetworkBalances.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runNetworkBalancesScraper(ctx, &svc)
		}()
	} else {
		svc.Log.Info("Not starting Network Balances Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Asset Prices
	//
	if svc.Config.Services.AssetPrices.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runAssetPricesScraper(ctx, &svc)
		}()
	} else {
		svc.Log.Info("Not starting Asset Prices Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Prometheus Endpoint
	//
	if svc.Config.Prometheus.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Starting Prometheus Endpoint service")
			if err := svc.PrometheusService.Start(); err != nil {
				svc.Log.Error("Failed to start Prometheus Endpoint", zap.Error(err))
				cancel()
			}
		}()
	} else {
		svc.Log.Info("Not starting Prometheus Endpoint", zap.String("config", "Enabled=false"))
	}
	//
	// start: DataNode Checker
	//
	if svc.Config.Prometheus.Enabled { // same flag as for Prometheus
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			svc.Log.Info("Starting DataNode Checker service in 10sec")
			time.Sleep(10 * time.Second)
			if err := svc.DataNodeCheckerService.Start(ctx); err != nil {
				svc.Log.Error("Failed to start DataNode Checker service", zap.Error(err))
				cancel()
			}
		}()
	} else {
		svc.Log.Info("Not starting DataNode Checker service", zap.String("config", "Enabled=false"))
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
		fmt.Printf("Evertything closed nicely\n")
	case <-time.After(5 * time.Second):
		fmt.Printf("Service timed out to stop. Force stopping\n")
	}

	//
	// DONE
	//
	time.Sleep(time.Millisecond * 100)
	fmt.Println("Service has stopped")
}

// Block Signers
func runBlockSignersScraper(ctx context.Context, svc *cmd.AllServices) {
	svc.Log.Info("Starting update Block Singers Scraper in 5sec")

	time.Sleep(5 * time.Second) // delay everything by 5sec

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runBlockSignerScrapper tick")

		if err := svc.UpdateService.UpdateBlockSignersAllNew(ctx); err != nil {
			svc.Log.Error("Failed to update Block Signers", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			svc.Log.Info("Stopping update Block Singers Scraper")
			return
		case <-ticker.C:
			continue
		}
	}
}

// Network History Segments
func runNetworkHistorySegmentsScraper(ctx context.Context, svc *cmd.AllServices) {
	svc.Log.Info("Starting update Network History Segments Scraper in 10sec")

	time.Sleep(10 * time.Second) // delay everything by 10sec

	ticker := time.NewTicker(250 * time.Second) // every ~300 block
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runNetworkHistorySegmentsScraper tick")

		apiURLs := []string{}
		for _, dataNode := range svc.Config.DataNode {
			apiURLs = append(apiURLs, dataNode.REST)
		}
		if err := svc.UpdateService.UpdateNetworkHistorySegments(ctx, apiURLs); err != nil {
			svc.Log.Error("Failed to update Network History Segments", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			svc.Log.Info("Stopping update Network History Segments Scraper")
			return
		case <-ticker.C:
			continue
		}
	}
}

// Comet Txs
func runCometTxsScraper(ctx context.Context, svc *cmd.AllServices) {
	svc.Log.Info("Starting update Comet Txs Scraper in 20sec")

	time.Sleep(20 * time.Second) // delay everything by 20sec - 15sec after Block Signers scraper

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runCometTxsScraper tick")

		if err := svc.UpdateService.UpdateCometTxsAllNew(ctx); err != nil {
			svc.Log.Error("Failed to update Comet Txs", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			svc.Log.Info("Stopping update Comet Txs Scraper")
			return
		case <-ticker.C:
			continue
		}
	}
}

// Network Balances
func runNetworkBalancesScraper(ctx context.Context, svc *cmd.AllServices) {
	svc.Log.Info("Starting update Network Balances Scraper in 15sec")

	time.Sleep(15 * time.Second)

	ticker := time.NewTicker(50 * time.Second)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runCometTxsScraper tick")

		if err := svc.UpdateService.UpdateAssetPoolBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Asset Pool", zap.Error(err))
		}
		if err := svc.UpdateService.UpdatePartiesTotalBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Partiest Total", zap.Error(err))
		}
		if err := svc.UpdateService.UpdateUnrealisedWithdrawalsBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Unrealised Withdrawals", zap.Error(err))
		}
		if err := svc.UpdateService.UpdateUnfinalizedDepositsBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Unfinalized Deposits", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			svc.Log.Info("Stopping update Network Balances Scraper")
			return
		case <-ticker.C:
			continue
		}
	}
}

// Asset Prices
func runAssetPricesScraper(ctx context.Context, svc *cmd.AllServices) {
	svc.Log.Info("Starting update Asset Prices Scraper in 25sec")

	time.Sleep(25 * time.Second)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		if err := svc.UpdateService.UpdateAssetPrices(ctx); err != nil {
			svc.Log.Error("Failed to update Asset Prices", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			svc.Log.Info("Stopping update Asset Prices Scraper")
			return
		case <-ticker.C:
			continue
		}
	}
}
