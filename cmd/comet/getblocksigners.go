package comet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
	"github.com/vegaprotocol/data-metrics-store/config"
)

type GetBlockSignersArgs struct {
	*CometArgs
	Block int
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

	getBlockSignersCmd.PersistentFlags().IntVar(&getBlockSignersArgs.Block, "block", 1, "Number of block to get data for")
	if err := getBlockSignersCmd.MarkPersistentFlagRequired("block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunGetBlockSigners(args GetBlockSignersArgs) error {

	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if len(args.ApiURL) == 0 {
		if cfg != nil {
			args.ApiURL = cfg.CometBFT.ApiURL
		} else {
			return fmt.Errorf("Required --api-url flag or config.toml file")
		}
	}

	client := comet.NewCometClient(args.ApiURL)

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
