package etherscan

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type EtherscanArgs struct {
	*rootCmd.RootArgs
	ApiURL string
	ApiKey string
}

var etherscanArgs EtherscanArgs

var EtherscanCmd = &cobra.Command{
	Use:   "etherscan",
	Short: "Interact with Etherscan",
	Long:  `Interact with Etherscan`,
}

func init() {
	etherscanArgs.RootArgs = &rootCmd.Args
	EtherscanCmd.PersistentFlags().StringVar(&etherscanArgs.ApiURL, "api-url", "", "Etherscan API URL")
	EtherscanCmd.PersistentFlags().StringVar(&etherscanArgs.ApiKey, "api-key", "", "Etherscan API Key")
}
