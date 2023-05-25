package main

import (
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
	"github.com/vegaprotocol/data-metrics-store/cmd/comet"
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.RootCmd.AddCommand(comet.CometCmd)
}
