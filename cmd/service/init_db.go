package service

import (
	"fmt"
	"os"

	"code.vegaprotocol.io/vega/logging"
	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
)

type InitDBArgs struct {
	*ServiceArgs
}

var initDBArgs InitDBArgs

// initDBCmd represents the initDB command
var initDBCmd = &cobra.Command{
	Use:   "init-db",
	Short: "Initiase monitoring schema in existing Data Node database",
	Long:  `Initiase monitoring schema in existing Data Node database`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunInitDB(initDBArgs); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ServiceCmd.AddCommand(initDBCmd)
	initDBArgs.ServiceArgs = &serviceArgs
}

func setupDB(logger *logging.Logger, cfg *config.Config) error {
	if err := sqlstore.MigrateToLatestSchema(logger, cfg.SQLStore.GetConnectionConfig()); err != nil {
		return fmt.Errorf("failed to migrate to the latest schema: %w", err)
	}

	if err := sqlstore.SetRetentionPolicies(cfg.SQLStore.GetConnectionConfig(), cfg.DataNodeDBExtension.BaseRetentionPolicy, cfg.DataNodeDBExtension.RetentionPolicy, logger); err != nil {
		return fmt.Errorf("failed to set retention policies: %w", err)
	}

	return nil
}

func RunInitDB(args InitDBArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return fmt.Errorf("failed to get config and logger: %w", err)
	}

	if err := setupDB(logger, cfg); err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}

	return nil
}
