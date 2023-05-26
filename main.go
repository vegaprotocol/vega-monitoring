package main

import (
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
	"github.com/vegaprotocol/data-metrics-store/cmd/comet"
	"github.com/vegaprotocol/data-metrics-store/cmd/etherscan"
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.RootCmd.AddCommand(comet.CometCmd)
	rootCmd.RootCmd.AddCommand(etherscan.EtherscanCmd)
}
