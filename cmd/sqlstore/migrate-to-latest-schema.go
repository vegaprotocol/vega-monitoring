package sqlstore

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/config"
	"github.com/vegaprotocol/data-metrics-store/sqlstore"
)

type MigrateToLatestSchemaArgs struct {
	*SQLStoreArgs
}

var migrateToLatestSchemaArgs MigrateToLatestSchemaArgs

// migrateToLatestSchemaCmd represents the migrateToLatestSchema command
var migrateToLatestSchemaCmd = &cobra.Command{
	Use:   "migrate-to-latest-schema",
	Short: "Migrate Monitoring Tables to Latest Goose Schema",
	Long:  `Migrate Monitoring Tables to Latest Goose Schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunMigrateToLatestSchema(migrateToLatestSchemaArgs); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	SQLStoreCmd.AddCommand(migrateToLatestSchemaCmd)
	migrateToLatestSchemaArgs.SQLStoreArgs = &sqlstoreArgs
}

func RunMigrateToLatestSchema(args MigrateToLatestSchemaArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := sqlstore.MigrateToLatestSchema(logger, cfg.GetConnectionConfig()); err != nil {
		return err
	}

	return nil
}
