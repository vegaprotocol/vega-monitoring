package etherscan

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type EtherscanArgs struct {
	*rootCmd.RootArgs
	EthNetwork string
	apiKey     string
}

var etherscanArgs EtherscanArgs

var EtherscanCmd = &cobra.Command{
	Use:   "etherscan",
	Short: "Interact with Etherscan",
	Long:  `Interact with Etherscan`,
}

func init() {
	etherscanArgs.RootArgs = &rootCmd.Args
	EtherscanCmd.PersistentFlags().StringVar(&etherscanArgs.EthNetwork, "eth-network", "mainnet", "Used with address, specify which Ethereum Network to use")
	EtherscanCmd.PersistentFlags().StringVar(&etherscanArgs.apiKey, "api-key", "", "Etherscan API Key")
}
