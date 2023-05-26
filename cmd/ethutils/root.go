package ethutils

import (
	"log"

	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type EthUtilsArgs struct {
	*rootCmd.RootArgs
	EthNodeURL string
}

var ethUtilsArgs EthUtilsArgs

var EthUtilsCmd = &cobra.Command{
	Use:   "eth-utils",
	Short: "Interact with Ethereum Node",
	Long:  `Interact with Ethereum Node`,
}

func init() {
	ethUtilsArgs.RootArgs = &rootCmd.Args
	EthUtilsCmd.PersistentFlags().StringVar(&ethUtilsArgs.EthNodeURL, "eth-node", "", "Ethereum Node URL")
	if err := EthUtilsCmd.MarkPersistentFlagRequired("eth-node"); err != nil {
		log.Fatalf("%v\n", err)
	}
}
