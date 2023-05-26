package binance

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/binance"
)

type GetPriceArgs struct {
	*BinanceArgs
	Asset string
}

var getPriceArgs GetPriceArgs

// getPriceCmd represents the getPrice command
var getPriceCmd = &cobra.Command{
	Use:   "get-price",
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
	BinanceCmd.AddCommand(getPriceCmd)
	getPriceArgs.BinanceArgs = &binanceArgs

	getPriceCmd.PersistentFlags().StringVar(&getPriceArgs.Asset, "wallet", "eth", "Ethereum address to get balance for")
}

func RunGetPrice(args GetPriceArgs) error {
	client := binance.NewBinanceClient(args.ApiURL)
	price, err := client.GetAssetPrice(args.Asset)
	if err != nil {
		return err
	}
	fmt.Printf("Price of '%s' in $usdt: %f $usdt\n", args.Asset, price)

	return nil
}
