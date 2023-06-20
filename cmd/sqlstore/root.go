package sqlstore

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type SQLStoreArgs struct {
	*rootCmd.RootArgs
}

var sqlstoreArgs SQLStoreArgs

var SQLStoreCmd = &cobra.Command{
	Use:   "sqlstore",
	Short: "Interact with Data Node TimescaleDB - extra monitoring tables",
	Long:  `Interact with Data Node TimescaleDB - extra monitoring tables`,
}

func init() {
	sqlstoreArgs.RootArgs = &rootCmd.Args
}
