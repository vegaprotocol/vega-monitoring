package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/cmd"
)

type CometTxsArgs struct {
	*UpdateArgs
	FromBlock int64
	ToBlock   int64
}

var cometTxsArgs CometTxsArgs

// getBlockCmd represents the getBlock command
var cometTxsCmd = &cobra.Command{
	Use:   "comet-txs",
	Short: "Get data from CometBFT API and store it in SQLStore",
	Long:  `Get data from CometBFT API and store it in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunCometTxs(cometTxsArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(cometTxsCmd)
	cometTxsArgs.UpdateArgs = &updateArgs

	cometTxsCmd.PersistentFlags().Int64Var(&cometTxsArgs.FromBlock, "from-block", 0, "First block to get")
	cometTxsCmd.PersistentFlags().Int64Var(&cometTxsArgs.ToBlock, "to-block", 0, "Last block to get")
}

func RunCometTxs(args CometTxsArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := svc.UpdateService.UpdateCometTxs(context.Background(), args.FromBlock, args.ToBlock); err != nil {
		return err
	}

	return nil
}
