package service

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type ServiceArgs struct {
	*rootCmd.RootArgs
	ConfigFilePath string
}

var serviceArgs ServiceArgs

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service to gather data and put it into the database",
	Long:  `Service to gather data and put it into the database`,
}

func init() {
	serviceArgs.RootArgs = &rootCmd.Args
	ServiceCmd.PersistentFlags().StringVar(&serviceArgs.ConfigFilePath, "config", "config.toml", "Path to the config file")
}
