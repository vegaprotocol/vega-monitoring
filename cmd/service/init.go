package service

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/config"
)

type InitArgs struct {
	*ServiceArgs
	Force bool
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
	initCmd.PersistentFlags().BoolVar(&initArgs.Force, "force", false, "Force creation config file, even it if exists")
}

func RunInit(args InitArgs) error {
	if !args.Force {
		if _, err := os.Stat(args.ConfigFilePath); err == nil {
			return fmt.Errorf("cannot create new config in %s. File aready exists. Use --force to overwrite.", args.ConfigFilePath)
		}
	}

	_, err := config.StoreDefaultConfigInFile(args.ConfigFilePath)
	if err != nil {
		return err
	}

	return nil
}
