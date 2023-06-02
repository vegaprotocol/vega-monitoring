package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/cmd"
)

type NetworkBalancesArgs struct {
	*UpdateArgs
	AssetPool bool
	All       bool
}

var networkBalancesArgs NetworkBalancesArgs

// getBlockCmd represents the getBlock command
var networkBalancesCmd = &cobra.Command{
	Use:   "network-balances",
	Short: "Update Network Balances in SQLStore",
	Long:  `Update Network Balances in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunNetworkBalances(networkBalancesArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(networkBalancesCmd)
	networkBalancesArgs.UpdateArgs = &updateArgs

	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.AssetPool, "asset-pool", false, "Update Asset Pool")
	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.All, "all", false, "Update all")
}

func RunNetworkBalances(args NetworkBalancesArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if args.All || args.AssetPool {
		if err := svc.UpdateService.UpdateAssetPoolBalances(context.Background()); err != nil {
			return err
		}
	}

	return nil
}
