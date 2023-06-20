package service

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/config"
	"github.com/vegaprotocol/data-metrics-store/sqlstore"
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

func RunInitDB(args InitDBArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := sqlstore.MigrateToLatestSchema(logger, cfg.SQLStore.GetConnectionConfig()); err != nil {
		return err
	}

	return nil
}
