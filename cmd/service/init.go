package service

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/config"
)

type InitArgs struct {
	*ServiceArgs
}

var initArgs InitArgs

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init config",
	Long:  `Init config`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunInit(initArgs); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ServiceCmd.AddCommand(initCmd)
	initArgs.ServiceArgs = &serviceArgs
}

func RunInit(args InitArgs) error {
	log, err := config.GetLogger(args.Debug)
	if err != nil {
		return err
	}

	_, err = config.StoreDefaultConfigInFile(args.ConfigFilePath, log)
	if err != nil {
		return err
	}

	return nil
}
