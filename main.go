package main

import (
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
	"github.com/vegaprotocol/data-metrics-store/cmd/binance"
	"github.com/vegaprotocol/data-metrics-store/cmd/coingecko"
	"github.com/vegaprotocol/data-metrics-store/cmd/comet"
	"github.com/vegaprotocol/data-metrics-store/cmd/etherscan"
	"github.com/vegaprotocol/data-metrics-store/cmd/ethutils"
	"github.com/vegaprotocol/data-metrics-store/cmd/service"
	"github.com/vegaprotocol/data-metrics-store/cmd/sqlstore"
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.RootCmd.AddCommand(comet.CometCmd)
	rootCmd.RootCmd.AddCommand(etherscan.EtherscanCmd)
	rootCmd.RootCmd.AddCommand(ethutils.EthUtilsCmd)
	rootCmd.RootCmd.AddCommand(binance.BinanceCmd)
	rootCmd.RootCmd.AddCommand(coingecko.CoingeckoCmd)
	rootCmd.RootCmd.AddCommand(service.ServiceCmd)
	rootCmd.RootCmd.AddCommand(sqlstore.SQLStoreCmd)
}
