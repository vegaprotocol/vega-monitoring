package coingecko

import (
	"github.com/spf13/cobra"
)

type CoingeckoArgs struct {
	ApiURL string
}

var coingeckoArgs CoingeckoArgs

var CoingeckoCmd = &cobra.Command{
	Use:   "coingecko",
	Short: "Interact with Coingecko API",
	Long:  `Interact with Coingecko API`,
}

func init() {
	CoingeckoCmd.PersistentFlags().StringVar(&coingeckoArgs.ApiURL, "api-url", "https://api.coingecko.com/api/v3", "Coingecko API URL")
}
