package service

import (
	"context"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vegaprotocol/vega-monitoring/cmd"
	"github.com/vegaprotocol/vega-monitoring/metamonitoring"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

const (
	BlockSignersLoopInterval    = 30 * time.Second
	NetworkHistoryLoopInterval  = 120 * time.Second
	CometTxsLoopInterval        = 20 * time.Second
	NetworkBalancesLoopInterval = 15 * time.Second
	AssetPricesLoopInterval     = 25 * time.Second
)

func startDataNodeDBExtension(
	svc *cmd.AllServices,
	shutdownWg *sync.WaitGroup,
	ctx context.Context,
) {
	if err := setupDB(svc.Log, svc.Config); err != nil {
		log.Fatalf("failed to setup database: %+v\n", err)
	}

	if err := sqlstore.MigrateToLatestSchema(svc.Log, svc.Config.SQLStore.GetConnectionConfig()); err != nil {
		log.Fatalf("failed to migrate database to latest version %+v\n", err)
	}

	//
	// start: Block Singers Service
	//
	if svc.Config.DataNodeDBExtension.BlockSigners.Enabled {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			runBlockSignersScraper(ctx, svc, svc.MonitoringService.BlockSignersStatusPublisher())
		}()
	} else {
		svc.Log.Info("Not starting Block Signers Service", zap.String("config", "Enabled=false"))
	}

	//
	// start: Network History Segments Service
	//
	if svc.Config.DataNodeDBExtension.NetworkHistorySegments.Enabled {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			runNetworkHistorySegmentsScraper(ctx, svc, svc.MonitoringService.SegmentsStatusPublisher())
		}()
	} else {
		svc.Log.Info("Not starting Network History Segments Service", zap.String("config", "Enabled=false"))
	}

	//
	// start: Comet Txs Service
	//
	if svc.Config.DataNodeDBExtension.CometTxs.Enabled {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			runCometTxsScraper(ctx, svc, svc.MonitoringService.CometTxsStatusPublisher())
		}()
	} else {
		svc.Log.Info("Not starting Comet Txs Service", zap.String("config", "Enabled=false"))
	}

	//
	// start: Network Balances
	//
	if svc.Config.DataNodeDBExtension.NetworkBalances.Enabled {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			runNetworkBalancesScraper(ctx, svc, svc.MonitoringService.NetworkBalancesStatusPublisher())
		}()
	} else {
		svc.Log.Info("Not starting Network Balances Service", zap.String("config", "Enabled=false"))
	}

	//
	// start: Asset Prices
	//
	if svc.Config.DataNodeDBExtension.AssetPrices.Enabled {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			runAssetPricesScraper(ctx, svc, svc.MonitoringService.AssetPricesStatusPublisher())
		}()
	} else {
		svc.Log.Info("Not starting Asset Prices Service", zap.String("config", "Enabled=false"))
	}

	//
	// start: Reporting the meta-monitoring statuses
	//
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		svc.MonitoringService.Run(ctx, NetworkHistoryLoopInterval*2)
	}()
}

// Block Signers
func runBlockSignersScraper(ctx context.Context, svc *cmd.AllServices, statusReporter metamonitoring.MonitoringStatusPublisher) {
	svc.Log.Info("Starting update Block Singers Scraper in 5sec")

	time.Sleep(5 * time.Second) // delay everything by 5sec

	ticker := time.NewTicker(BlockSignersLoopInterval)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runBlockSignerScrapper tick")

		if err := svc.UpdateService.UpdateBlockSignersAllNew(ctx); err != nil {
			if err := statusReporter.Publish(false); err != nil {
				svc.Log.Error("failed to publish false health check for the block signers svc", zap.Error(err))
			}
			svc.Log.Error("Failed to update Block Signers", zap.Error(err))
		} else {
			if err := statusReporter.Publish(true); err != nil {
				svc.Log.Error("failed to publish true health check for the block signers svc", zap.Error(err))
			}
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
func runNetworkHistorySegmentsScraper(ctx context.Context, svc *cmd.AllServices, statusReporter metamonitoring.MonitoringStatusPublisher) {
	svc.Log.Info("Starting update Network History Segments Scraper in 10sec")

	time.Sleep(10 * time.Second) // delay everything by 10sec

	ticker := time.NewTicker(NetworkHistoryLoopInterval) // every ~300 block
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runNetworkHistorySegmentsScraper tick")

		apiURLs := []string{}
		for _, dataNode := range svc.Config.Monitoring.DataNode {
			apiURLs = append(apiURLs, dataNode.REST)
		}
		if err := svc.UpdateService.UpdateNetworkHistorySegments(ctx, apiURLs); err != nil {
			svc.Log.Error("Failed to update Network History Segments", zap.Error(err))
			if err := statusReporter.Publish(false); err != nil {
				svc.Log.Error("failed to publish false health check for the network history segments svc", zap.Error(err))
			}
		} else {
			if err := statusReporter.Publish(true); err != nil {
				svc.Log.Error("failed to publish true health check for the network history segments svc", zap.Error(err))
			}
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
func runCometTxsScraper(ctx context.Context, svc *cmd.AllServices, statusReporter metamonitoring.MonitoringStatusPublisher) {
	svc.Log.Info("Starting update Comet Txs Scraper in 20sec")

	time.Sleep(20 * time.Second) // delay everything by 20sec - 15sec after Block Signers scraper

	ticker := time.NewTicker(CometTxsLoopInterval)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runCometTxsScraper tick")

		if err := svc.UpdateService.UpdateCometTxsAllNew(ctx); err != nil {
			svc.Log.Error("Failed to update Comet Txs", zap.Error(err))
			if err := statusReporter.Publish(false); err != nil {
				svc.Log.Error("failed to publish false health check for the comet txs svc", zap.Error(err))
			}
		} else {
			if err := statusReporter.Publish(true); err != nil {
				svc.Log.Error("failed to publish true health check for the comet txs svc", zap.Error(err))
			}
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
func runNetworkBalancesScraper(ctx context.Context, svc *cmd.AllServices, statusReporter metamonitoring.MonitoringStatusPublisher) {
	svc.Log.Info("Starting update Network Balances Scraper in 15sec")

	time.Sleep(NetworkBalancesLoopInterval)

	ticker := time.NewTicker(50 * time.Second)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runNetworkBalancesScraper tick")

		success := true
		if err := svc.UpdateService.UpdateAssetPoolBalances(ctx, svc.Config.Ethereum.AssetPoolAddress); err != nil {
			svc.Log.Error("Failed to update Network Balances: Asset Pool", zap.Error(err))
			success = false
		}
		if err := svc.UpdateService.UpdatePartiesTotalBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Partiest Total", zap.Error(err))
			success = false
		}
		if err := svc.UpdateService.UpdateUnrealisedWithdrawalsBalances(ctx); err != nil {
			svc.Log.Error(
				"Failed to update Network Balances: Unrealised Withdrawals",
				zap.Error(err),
			)

			success = false
		}
		if err := svc.UpdateService.UpdateUnfinalizedDepositsBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Unfinalized Deposits", zap.Error(err))
			success = false
		}

		if err := statusReporter.Publish(success); err != nil {
			svc.Log.Error("failed to publish %v health check for the network balance svc", zap.Error(err))
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
func runAssetPricesScraper(ctx context.Context, svc *cmd.AllServices, statusReporter metamonitoring.MonitoringStatusPublisher) {
	svc.Log.Info("Starting update Asset Prices Scraper in 25sec")

	time.Sleep(25 * time.Second)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		svc.Log.Debugf("runAssetPricesScraper tick")

		if err := svc.UpdateService.UpdateAssetPrices(ctx); err != nil {
			svc.Log.Error("Failed to update Asset Prices", zap.Error(err))

			if err := statusReporter.Publish(false); err != nil {
				svc.Log.Error("failed to publish false health check for the asset price svc", zap.Error(err))
			}
		} else {
			if err := statusReporter.Publish(true); err != nil {
				svc.Log.Error("failed to publish false health check for the asset price svc", zap.Error(err))
			}
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
