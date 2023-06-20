package sqlstore

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/config"
	"github.com/vegaprotocol/data-metrics-store/sqlstore"
)

type RemoveSchemaArgs struct {
	*SQLStoreArgs
}

var removeSchemaArgs RemoveSchemaArgs

// removeSchemaCmd represents the removeSchema command
var removeSchemaCmd = &cobra.Command{
	Use:   "remove-schema",
	Short: "Revert Monitoring Tables to SQL schema to version 0",
	Long:  `Revert Monitoring Tables to SQL schema to version 0`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunRemoveSchema(removeSchemaArgs); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	SQLStoreCmd.AddCommand(removeSchemaCmd)
	removeSchemaArgs.SQLStoreArgs = &sqlstoreArgs
}

func RunRemoveSchema(args RemoveSchemaArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	if err := sqlstore.RevertToSchemaVersionZero(logger, cfg.SQLStore.GetConnectionConfig()); err != nil {
		return err
	}

	return nil
}
