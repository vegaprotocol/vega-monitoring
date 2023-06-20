package update

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/cmd"
)

type AssetPricesArgs struct {
	*UpdateArgs
}

var assetPricesArgs AssetPricesArgs

var assetPricesCmd = &cobra.Command{
	Use:   "asset-prices",
	Short: "Get data from Coingecko REST API about prices of tokens from config.toml and store it in SQLStore",
	Long:  `Get data from Coingecko REST API about prices of tokens from config.toml and store it in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunAssetPrices(assetPricesArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(assetPricesCmd)
	assetPricesArgs.UpdateArgs = &updateArgs
}

func RunAssetPrices(args AssetPricesArgs) error {
	svc, err := cmd.SetupServices(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := svc.UpdateService.UpdateAssetPrices(context.Background()); err != nil {
		return err
	}

	return nil
}
