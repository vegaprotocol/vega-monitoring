package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/cmd"
)

type BlockSignersArgs struct {
	*UpdateArgs
	FromBlock int64
	ToBlock   int64
}

var blockSignersArgs BlockSignersArgs

// getBlockCmd represents the getBlock command
var blockSignersCmd = &cobra.Command{
	Use:   "block-signers",
	Short: "Get data from CometBFT REST API and store it in SQLStore",
	Long:  `Get data from CometBFT REST API and store it in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunBlockSigners(blockSignersArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(blockSignersCmd)
	blockSignersArgs.UpdateArgs = &updateArgs

	blockSignersCmd.PersistentFlags().Int64Var(&blockSignersArgs.FromBlock, "from-block", 0, "First block to get")
	blockSignersCmd.PersistentFlags().Int64Var(&blockSignersArgs.ToBlock, "to-block", 0, "Last block to get")
}

func RunBlockSigners(args BlockSignersArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := svc.UpdateService.UpdateBlockSigners(context.Background(), args.FromBlock, args.ToBlock); err != nil {
		return err
	}

	return nil
}
