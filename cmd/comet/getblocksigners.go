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

type GetBlockSignersArgs struct {
	*CometArgs
	Block int64
}

var getBlockSignersArgs GetBlockSignersArgs

// getBlockSignersCmd represents the getBlockSigners command
var getBlockSignersCmd = &cobra.Command{
	Use:   "get-block-signers",
	Short: "Get BlockSigners Data from CometBFT",
	Long:  `Get BlockSigners Data from CometBFT`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetBlockSigners(getBlockSignersArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	CometCmd.AddCommand(getBlockSignersCmd)
	getBlockSignersArgs.CometArgs = &cometArgs

	getBlockSignersCmd.PersistentFlags().Int64Var(&getBlockSignersArgs.Block, "block", 1, "Number of block to get data for")
	if err := getBlockSignersCmd.MarkPersistentFlagRequired("block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunGetBlockSigners(args GetBlockSignersArgs) error {

	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)

	var client *comet.CometClient
	if len(args.ApiURL) > 0 {
		client = comet.NewCometClient(&config.CometBFTConfig{ApiURL: args.ApiURL})
	} else if cfg != nil {
		client = comet.NewCometClient(&cfg.CometBFT)
	} else {
		return fmt.Errorf("Required --api-url flag or config.toml file")
	}

	blockSignersData, err := client.GetBlockSigners(args.Block)
	if err != nil {
		return err
	}

	byteBlockSignersData, err := json.MarshalIndent(blockSignersData, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse blockSigners data for blockSigners '%d' got from %s, %w", args.Block, args.ApiURL, err)
	}
	fmt.Println(string(byteBlockSignersData))

	return nil
}
