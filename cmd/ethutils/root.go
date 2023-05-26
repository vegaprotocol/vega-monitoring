package ethutils

import (
	"log"

	"github.com/spf13/cobra"
)

type EthUtilsArgs struct {
	EthNodeURL string
}

var ethUtilsArgs EthUtilsArgs

var EthUtilsCmd = &cobra.Command{
	Use:   "eth-utils",
	Short: "Interact with Ethereum Node",
	Long:  `Interact with Ethereum Node`,
}

func init() {
	EthUtilsCmd.PersistentFlags().StringVar(&ethUtilsArgs.EthNodeURL, "eth-node", "", "Ethereum Node URL")
	if err := EthUtilsCmd.MarkPersistentFlagRequired("eth-node"); err != nil {
		log.Fatalf("%v\n", err)
	}
}
