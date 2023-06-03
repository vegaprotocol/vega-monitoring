package coingecko

import (
	"encoding/json"
	"fmt"
	"os"

	"code.vegaprotocol.io/vega/logging"
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
	cfg, log, _ := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)

	coingeckoConfig := config.CoingeckoConfig{}
	if len(args.ApiURL) > 0 {
		coingeckoConfig.ApiURL = args.ApiURL
	} else if cfg != nil {
		coingeckoConfig.ApiURL = cfg.Coingecko.ApiURL
	} else {
		return fmt.Errorf("Required --api-url flag or config.toml file")
	}

	if log == nil {
		log = logging.NewProdLogger()
	}

	client := coingecko.NewCoingeckoClient(&coingeckoConfig, log)
	var prices []coingecko.PriceData
	var err error

	if args.AssetIds != nil && len(args.AssetIds) > 0 {
		prices, err = client.GetAssetPrices(args.AssetIds)
		if err != nil {
			return err
		}
	} else if cfg != nil {
		coingeckoConfig.AssetIds = cfg.Coingecko.AssetIds
		prices, err = client.GetPrices()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Required --asset-ids flag or config.toml file")
	}

	bytePrices, err := json.MarshalIndent(prices, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse prices to json %v, %w", prices, err)
	}
	fmt.Println(string(bytePrices))

	return nil
}
