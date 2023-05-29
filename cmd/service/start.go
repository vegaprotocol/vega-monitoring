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
		// main
		_, cancel := startEverything(startArgs)
		defer shutdownEverything(cancel)
		waitShutdown()
	},
}

func init() {
	ServiceCmd.AddCommand(startCmd)
	startArgs.ServiceArgs = &serviceArgs
}

func startEverything(args StartArgs) (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithCancel(context.Background())
	//
	// Setup your services here and start all the go routines below
	//
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		log.Fatalf("Failed to setup Services %+v\n", err)
	}
	go func() {
		svc.Log.Info("Starting update Block Singers service")
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		time.Sleep(5 * time.Second)
		for {
			if err := svc.UpdateService.UpdateBlockSignersAllNew(); err != nil {
				svc.Log.Error("Failed to update Block Signers", zap.Error(err))
			}

			select {
			case <-ctx.Done():
				svc.Log.Info("Stopping update Block Singers service")
				return
			case <-ticker.C:
				continue
			}
		}
	}()
	go func() {
		fmt.Printf("Starting something #2\n")
	}()

	fmt.Printf("Service has started\n")
	return
}

func shutdownEverything(cancel context.CancelFunc) {
	cancel()

	var wg sync.WaitGroup
	//
	// Shut down all your services here - in go routines
	//
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Shutting down something #1\n")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Shutting down something #2\n")
	}()

	// Notify when all go routines are done
	c := make(chan struct{})
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()

	// Timeout
	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelTimeout()

	// Wait for go routines or timeout
	select {
	case <-c:
		fmt.Printf("Evertything closed nicely\n")
	case <-ctx.Done():
		fmt.Printf("Service timed out to stop. Force stopping\n")
	}

	time.Sleep(time.Millisecond * 100)
	fmt.Println("Service has stopped")
}

func waitShutdown() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigc
	log.Printf("signal received [%v] shutting down\n", s)
}
