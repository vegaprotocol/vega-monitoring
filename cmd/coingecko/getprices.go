package coingecko

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/coingecko"
	"github.com/vegaprotocol/data-metrics-store/config"
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

	getPriceCmd.PersistentFlags().StringSliceVar(&getPriceArgs.AssetIds, "asset-ids", nil, "Comma separated list of Coingecko's Asset ids")
}

func RunGetPrice(args GetPriceArgs) error {
	cfg, _, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if len(args.ApiURL) == 0 {
		if cfg != nil {
			args.ApiURL = cfg.Coingecko.ApiURL
		} else {
			return fmt.Errorf("Required --api-url flag or config.toml file")
		}
	}
	if args.AssetIds == nil || len(args.AssetIds) == 0 {
		if cfg != nil {
			args.AssetIds = []string{}
			for _, assetId := range cfg.Coingecko.AssetIds {
				args.AssetIds = append(args.AssetIds, assetId)
			}
		} else {
			return fmt.Errorf("Required --asset-ids flag or config.toml file")
		}
	}

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
