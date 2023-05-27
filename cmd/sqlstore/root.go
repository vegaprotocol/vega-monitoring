package sqlstore

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type SQLStoreArgs struct {
	*rootCmd.RootArgs
	ConfigFilePath string
}

var sqlstoreArgs SQLStoreArgs

var SQLStoreCmd = &cobra.Command{
	Use:   "sqlstore",
	Short: "Interact with Data Node TimescaleDB - extra monitoring tables",
	Long:  `Interact with Data Node TimescaleDB - extra monitoring tables`,
}

func init() {
	sqlstoreArgs.RootArgs = &rootCmd.Args
	SQLStoreCmd.PersistentFlags().StringVar(&sqlstoreArgs.ConfigFilePath, "config", "config.toml", "Path to the config file")
}
