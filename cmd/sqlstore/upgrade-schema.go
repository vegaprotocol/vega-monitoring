package sqlstore

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

type UpgradeSchemaArgs struct {
	*SQLStoreArgs
}

var upgradeSchemaArgs UpgradeSchemaArgs

// upgradeSchemaCmd represents the upgradeSchema command
var upgradeSchemaCmd = &cobra.Command{
	Use:   "upgrade-schema",
	Short: "Migrate Monitoring Tables to Latest Goose Schema",
	Long:  `Migrate Monitoring Tables to Latest Goose Schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunUpgradeSchema(upgradeSchemaArgs); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	SQLStoreCmd.AddCommand(upgradeSchemaCmd)
	upgradeSchemaArgs.SQLStoreArgs = &sqlstoreArgs
}

func RunUpgradeSchema(args UpgradeSchemaArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := sqlstore.MigrateToLatestSchema(logger, cfg.SQLStore.GetConnectionConfig()); err != nil {
		return err
	}

	return nil
}
