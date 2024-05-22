package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vegaprotocol/vega-monitoring/cmd"
)

type NetworkBalancesArgs struct {
	*UpdateArgs
	All                   bool
	AssetPool             bool
	PartiesTotal          bool
	UnrealisedWithdrawals bool
	UnfinalizedDeposits   bool
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

	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.All, "all", false, "Update all Balances")
	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.AssetPool, "asset-pool", false, "Update Asset Pool Balances")
	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.PartiesTotal, "parties-total", false, "Update Parties Total Balances")
	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.UnrealisedWithdrawals, "unrealised-withdrawals", false, "Update Unrealised Withdrawals Balances")
	networkBalancesCmd.PersistentFlags().BoolVar(&networkBalancesArgs.UnfinalizedDeposits, "unfinalized-deposits", false, "Update Unfinalized Deposits Balances")
}

func RunNetworkBalances(args NetworkBalancesArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if args.All || args.AssetPool {
		if err := svc.UpdateService.UpdateAssetPoolBalances(context.Background(), svc.Config.Ethereum, svc.Config.Arbitrum); err != nil {
			return err
		}
	}

	if args.All || args.PartiesTotal {
		if err := svc.UpdateService.UpdatePartiesTotalBalances(context.Background()); err != nil {
			return err
		}
	}

	if args.All || args.UnrealisedWithdrawals {
		if err := svc.UpdateService.UpdateUnrealisedWithdrawalsBalances(context.Background()); err != nil {
			return err
		}
	}

	if args.All || args.UnfinalizedDeposits {
		if err := svc.UpdateService.UpdateUnfinalizedDepositsBalances(context.Background()); err != nil {
			return err
		}
	}

	return nil
}
