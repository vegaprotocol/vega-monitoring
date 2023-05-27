package coingecko

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type CoingeckoArgs struct {
	*rootCmd.RootArgs
	ApiURL string
}

var coingeckoArgs CoingeckoArgs

var CoingeckoCmd = &cobra.Command{
	Use:   "coingecko",
	Short: "Interact with Coingecko API",
	Long:  `Interact with Coingecko API`,
}

func init() {
	coingeckoArgs.RootArgs = &rootCmd.Args
	CoingeckoCmd.PersistentFlags().StringVar(&coingeckoArgs.ApiURL, "api-url", "", "Coingecko API URL")
}
