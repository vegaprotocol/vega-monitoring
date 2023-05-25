package comet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/comet"
)

type GetBlockSignersRangeArgs struct {
	*CometArgs
	FromBlock int
	ToBlock   int
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

	getBlockSignersRangeCmd.PersistentFlags().IntVar(&getBlockSignersRangeArgs.FromBlock, "from-block", 1, "First block to get")
	if err := getBlockSignersRangeCmd.MarkPersistentFlagRequired("from-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
	getBlockSignersRangeCmd.PersistentFlags().IntVar(&getBlockSignersRangeArgs.ToBlock, "to-block", 1, "Last block to get")
	if err := getBlockSignersRangeCmd.MarkPersistentFlagRequired("to-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunGetBlockSignersRange(args GetBlockSignersRangeArgs) error {
	client := comet.NewCometClient(args.ApiURL)

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
