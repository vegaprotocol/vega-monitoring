package main

import (
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
	"github.com/vegaprotocol/vega-monitoring/cmd/binance"
	"github.com/vegaprotocol/vega-monitoring/cmd/coingecko"
	"github.com/vegaprotocol/vega-monitoring/cmd/comet"
	"github.com/vegaprotocol/vega-monitoring/cmd/etherscan"
	"github.com/vegaprotocol/vega-monitoring/cmd/ethutils"
	"github.com/vegaprotocol/vega-monitoring/cmd/healthcheck"
	"github.com/vegaprotocol/vega-monitoring/cmd/service"
	"github.com/vegaprotocol/vega-monitoring/cmd/sqlstore"
	"github.com/vegaprotocol/vega-monitoring/cmd/update"
	"github.com/vegaprotocol/vega-monitoring/cmd/version"
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
	rootCmd.RootCmd.AddCommand(update.UpdateCmd)
	rootCmd.RootCmd.AddCommand(version.VersionCmd)
	rootCmd.RootCmd.AddCommand(healthcheck.HealthcheckCmd)
}
