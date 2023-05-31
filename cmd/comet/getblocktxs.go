package comet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
	"github.com/vegaprotocol/data-metrics-store/config"
)

type GetBlockTxsArgs struct {
	*CometArgs
	Block int64
}

var getBlockTxsArgs GetBlockTxsArgs

// getBlockTxsCmd represents the getBlockTxs command
var getBlockTxsCmd = &cobra.Command{
	Use:   "get-block-txs",
	Short: "Get Txs for Block from CometBFT",
	Long:  `Get Txs for Block from CometBFT`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetBlockTxs(getBlockTxsArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	CometCmd.AddCommand(getBlockTxsCmd)
	getBlockTxsArgs.CometArgs = &cometArgs

	getBlockTxsCmd.PersistentFlags().Int64Var(&getBlockTxsArgs.Block, "block", 0, "Number of block to get data for (0 last block)")
}

func RunGetBlockTxs(args GetBlockTxsArgs) error {
	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)

	var client *comet.CometClient
	if len(args.ApiURL) > 0 {
		client = comet.NewCometClient(&config.CometBFTConfig{ApiURL: args.ApiURL})
	} else if cfg != nil {
		client = comet.NewCometClient(&cfg.CometBFT)
	} else {
		return fmt.Errorf("Required --api-url flag or config.toml file")
	}

	txsList, err := client.GetTxsForBlock(args.Block)
	if err != nil {
		return err
	}

	byteTxsList, err := json.MarshalIndent(txsList, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse Txs for block '%d' got from %s, %w", args.Block, args.ApiURL, err)
	}
	fmt.Println(string(byteTxsList))

	return nil
}
