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

type GetBlockSignersRangeArgs struct {
	*CometArgs
	FromBlock int64
	ToBlock   int64
}

var getBlockSignersRangeArgs GetBlockSignersRangeArgs

// getBlockCmd represents the getBlock command
var getBlockSignersRangeCmd = &cobra.Command{
	Use:   "get-block-signers-range",
	Short: "Get range of Block Data from CometBFT",
	Long:  `Get range of Block Data from CometBFT`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetBlockSignersRange(getBlockSignersRangeArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	CometCmd.AddCommand(getBlockSignersRangeCmd)
	getBlockSignersRangeArgs.CometArgs = &cometArgs

	getBlockSignersRangeCmd.PersistentFlags().Int64Var(&getBlockSignersRangeArgs.FromBlock, "from-block", 1, "First block to get")
	if err := getBlockSignersRangeCmd.MarkPersistentFlagRequired("from-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
	getBlockSignersRangeCmd.PersistentFlags().Int64Var(&getBlockSignersRangeArgs.ToBlock, "to-block", 1, "Last block to get")
	if err := getBlockSignersRangeCmd.MarkPersistentFlagRequired("to-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunGetBlockSignersRange(args GetBlockSignersRangeArgs) error {
	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	var client *comet.CometClient
	if len(args.ApiURL) > 0 {
		client = comet.NewCometClient(&config.LocalNodeConfig{CometURL: args.ApiURL})
	} else if cfg != nil {
		client = comet.NewCometClient(&cfg.LocalNode)
	} else {
		return fmt.Errorf("Required --api-url flag or config.toml file")
	}

	blockDataList, err := client.GetBlockSignersRange(args.FromBlock, args.ToBlock)
	if err != nil {
		return err
	}

	byteBlockDataList, err := json.MarshalIndent(blockDataList, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse block data for blocks from '%d' to '%d' got from %s, %w", args.FromBlock, args.ToBlock, args.ApiURL, err)
	}
	fmt.Println(string(byteBlockDataList))

	return nil
}
