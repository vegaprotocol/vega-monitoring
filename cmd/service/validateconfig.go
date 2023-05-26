package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/config"
)

type ValidateConfigArgs struct {
	*ServiceArgs
	Print bool
}

var validateConfigArgs ValidateConfigArgs

// validateConfigCmd represents the validateConfig command
var validateConfigCmd = &cobra.Command{
	Use:   "validate-config",
	Short: "Get Balance for an account for a token",
	Long:  `Get Balance for an account for a token`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunValidateConfig(validateConfigArgs); err != nil {
			fmt.Printf("Config Validation Failed:\n%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ServiceCmd.AddCommand(validateConfigCmd)
	validateConfigArgs.ServiceArgs = &serviceArgs
	validateConfigCmd.PersistentFlags().BoolVar(&validateConfigArgs.Print, "print", false, "Print out the config after reading it")
}

func RunValidateConfig(args ValidateConfigArgs) error {
	log, err := config.GetLogger(args.Debug)
	if err != nil {
		return err
	}
	cfg, err := config.ReadConfigAndWatch(args.ConfigFilePath, log)
	if err != nil {
		return fmt.Errorf("failed to read config file, %w", err)
	}
	if args.Print {
		byteCfg, err := json.MarshalIndent(cfg, "", "\t")
		if err != nil {
			return fmt.Errorf("failed to parse config from file %s, %w", args.ConfigFilePath, err)
		}
		fmt.Println(string(byteCfg))
	}

	return nil
}
