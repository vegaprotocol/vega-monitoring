package binance

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type BinanceArgs struct {
	*rootCmd.RootArgs
	ApiURL string
}

var binanceArgs BinanceArgs

var BinanceCmd = &cobra.Command{
	Use:   "binance",
	Short: "Interact with Binance API",
	Long:  `Interact with Binance API`,
}

func init() {
	binanceArgs.RootArgs = &rootCmd.Args
	BinanceCmd.PersistentFlags().StringVar(&binanceArgs.ApiURL, "api-url", "https://api.binance.com/api/v3", "Binance API URL")
}
