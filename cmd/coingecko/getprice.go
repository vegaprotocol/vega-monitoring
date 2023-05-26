package coingecko

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/coingecko"
)

type GetPriceArgs struct {
	*CoingeckoArgs
	AssetIds []string
}

var getPriceArgs GetPriceArgs

// getPriceCmd represents the getPrice command
var getPriceCmd = &cobra.Command{
	Use:   "get-prices",
	Short: "Get Price of Token in USDT",
	Long:  `Get Price of Token in USDT`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetPrice(getPriceArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	CoingeckoCmd.AddCommand(getPriceCmd)
	getPriceArgs.CoingeckoArgs = &coingeckoArgs

	getPriceCmd.PersistentFlags().StringSliceVar(&getPriceArgs.AssetIds, "wallet", []string{"vega-protocol", "tether", "usd-coin"}, "Comma separated list of Coingecko's Asset ids")
}

func RunGetPrice(args GetPriceArgs) error {
	client := coingecko.NewCoingeckoClient(args.ApiURL)
	prices, err := client.GetAssetPrices(args.AssetIds)
	if err != nil {
		return err
	}

	bytePrices, err := json.MarshalIndent(prices, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse prices to json %v, %w", prices, err)
	}
	fmt.Println(string(bytePrices))

	return nil
}
