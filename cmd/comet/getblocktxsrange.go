package comet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/clients/comet"
	"github.com/vegaprotocol/vega-monitoring/config"
)

type GetBlockTxsRangeArgs struct {
	*CometArgs
	FromBlock int64
	ToBlock   int64
}

var getBlockTxsRangeArgs GetBlockTxsRangeArgs

// getBlockTxsRangeCmd represents the getBlockTxsRange command
var getBlockTxsRangeCmd = &cobra.Command{
	Use:   "get-block-txs-range",
	Short: "Get Txs for Block from CometBFT",
	Long:  `Get Txs for Block from CometBFT`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetBlockTxsRange(getBlockTxsRangeArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	CometCmd.AddCommand(getBlockTxsRangeCmd)
	getBlockTxsRangeArgs.CometArgs = &cometArgs

	getBlockTxsRangeCmd.PersistentFlags().Int64Var(&getBlockTxsRangeArgs.FromBlock, "from-block", 1, "First block to get")
	if err := getBlockTxsRangeCmd.MarkPersistentFlagRequired("from-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
	getBlockTxsRangeCmd.PersistentFlags().Int64Var(&getBlockTxsRangeArgs.ToBlock, "to-block", 1, "Last block to get")
	if err := getBlockTxsRangeCmd.MarkPersistentFlagRequired("to-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunGetBlockTxsRange(args GetBlockTxsRangeArgs) error {
	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)

	var client *comet.CometClient
	if len(args.ApiURL) > 0 {
		client = comet.NewCometClient(&config.CometBFTConfig{ApiURL: args.ApiURL})
	} else if cfg != nil {
		client = comet.NewCometClient(&cfg.CometBFT)
	} else {
		return fmt.Errorf("Required --api-url flag or config.toml file")
	}

	txsList, err := client.GetTxsForBlockRange(args.FromBlock, args.ToBlock)
	if err != nil {
		return err
	}

	byteTxsList, err := json.MarshalIndent(txsList, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse Txs for block range from '%d' to '%d', got from %s, %w", args.FromBlock, args.ToBlock, args.ApiURL, err)
	}
	fmt.Println(string(byteTxsList))

	return nil
}
