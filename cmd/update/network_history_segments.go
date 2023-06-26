package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/cmd"
)

type NetworkHistorySegmentsArgs struct {
	*UpdateArgs
	ApiURL string
}

var networkHistorySegmentsArgs NetworkHistorySegmentsArgs

// getBlockCmd represents the getBlock command
var networkHistorySegmentsCmd = &cobra.Command{
	Use:   "network-history-segments",
	Short: "Get data from DataNode REST API and store it in SQLStore",
	Long:  `Get data from DataNode REST API and store it in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunNetworkHistorySegments(networkHistorySegmentsArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(networkHistorySegmentsCmd)
	networkHistorySegmentsArgs.UpdateArgs = &updateArgs

	networkHistorySegmentsCmd.PersistentFlags().StringVar(&networkHistorySegmentsArgs.ApiURL, "api-url", "", "Data Node URL")
}

func RunNetworkHistorySegments(args NetworkHistorySegmentsArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	apiURLs := []string{}

	if len(args.ApiURL) == 0 {
		for _, dataNode := range svc.Config.Monitoring.DataNode {
			apiURLs = append(apiURLs, dataNode.REST)
		}
	} else {
		apiURLs = append(apiURLs, args.ApiURL)
	}

	if err := svc.UpdateService.UpdateNetworkHistorySegments(context.Background(), apiURLs); err != nil {
		return err
	}

	return nil
}
