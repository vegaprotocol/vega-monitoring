package service

import (
	"context"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vegaprotocol/vega-monitoring/cmd"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

func startDataNodeDBExtension(
	svc *cmd.AllServices,
	shutdown_wg *sync.WaitGroup,
	ctx context.Context,
) {
	if err := sqlstore.MigrateToLatestSchema(svc.Log, svc.Config.SQLStore.GetConnectionConfig()); err != nil {
		log.Fatalf("Failed to migrate database to latest version %+v\n", err)
	}

	//
	// start: Block Singers Service
	//
	if svc.Config.DataNodeDBExtension.BlockSigners.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runBlockSignersScraper(ctx, svc)
		}()
	} else {
		svc.Log.Info("Not starting Block Signers Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Network History Segments Service
	//
	if svc.Config.DataNodeDBExtension.NetworkHistorySegments.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runNetworkHistorySegmentsScraper(ctx, svc)
		}()
	} else {
		svc.Log.Info("Not starting Network History Segments Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Comet Txs Service
	//
	if svc.Config.DataNodeDBExtension.CometTxs.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runCometTxsScraper(ctx, svc)
		}()
	} else {
		svc.Log.Info("Not starting Comet Txs Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Network Balances
	//
	if svc.Config.DataNodeDBExtension.NetworkBalances.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runNetworkBalancesScraper(ctx, svc)
		}()
	} else {
		svc.Log.Info("Not starting Network Balances Service", zap.String("config", "Enabled=false"))
	}
	//
	// start: Asset Prices
	//
	if svc.Config.DataNodeDBExtension.AssetPrices.Enabled {
		shutdown_wg.Add(1)
		go func() {
			defer shutdown_wg.Done()
			runAssetPricesScraper(ctx, svc)
		}()
	} else {
		svc.Log.Info("Not starting Asset Prices Service", zap.String("config", "Enabled=false"))
	}
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
		for _, dataNode := range svc.Config.Monitoring.DataNode {
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
		svc.Log.Debugf("runNetworkBalancesScraper tick")

		if err := svc.UpdateService.UpdateAssetPoolBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Asset Pool", zap.Error(err))
		}
		if err := svc.UpdateService.UpdatePartiesTotalBalances(ctx); err != nil {
			svc.Log.Error("Failed to update Network Balances: Partiest Total", zap.Error(err))
		}
		if err := svc.UpdateService.UpdateUnrealisedWithdrawalsBalances(ctx); err != nil {
			svc.Log.Error(
				"Failed to update Network Balances: Unrealised Withdrawals",
				zap.Error(err),
			)
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
		svc.Log.Debugf("runAssetPricesScraper tick")

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
