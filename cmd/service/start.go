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
	"github.com/vegaprotocol/data-metrics-store/cmd"
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

	//
	// start: Block Singers Service
	//
	shutdown_wg.Add(1)
	go func() {
		defer shutdown_wg.Done()
		runBlockSignersScraper(ctx, &svc)
	}()
	//
	// start: Network History Segments Service
	//
	shutdown_wg.Add(1)
	go func() {
		defer shutdown_wg.Done()
		runNetworkHistorySegmentsScraper(ctx, &svc)
	}()
	//
	// start: example service
	//
	// shutdown_wg.Add(1)
	// go func() {
	// 	defer shutdown_wg.Done()
	// 	fmt.Printf("Starting something #2\n")
	// }()

	fmt.Printf("Service has started\n")

	//
	// wait: For SIGNALL
	//
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigc
	log.Printf("signal received [%v] shutting down\n", s)

	//
	// Send CANCEL to all services
	//
	cancel()

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
	svc.Log.Info("Starting update Network History Segments Scraper in 15sec")

	time.Sleep(15 * time.Second) // delay everything by 5sec

	ticker := time.NewTicker(250 * time.Second) // every ~300 block
	defer ticker.Stop()

	for {
		apiURLs := []string{}
		for _, apiURL := range svc.Config.DataNode.Monitor {
			apiURLs = append(apiURLs, apiURL)
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
