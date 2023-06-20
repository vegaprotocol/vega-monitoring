package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RootArgs struct {
	Debug          bool
	ConfigFilePath string
}

var Args RootArgs

var RootCmd = &cobra.Command{
	Use:   "vega-monitoring",
	Short: "",
	Long:  ``,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run vega-monitoring '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.PersistentFlags().BoolVar(&Args.Debug, "debug", false, "Print debug logs")
	RootCmd.PersistentFlags().StringVar(&Args.ConfigFilePath, "config", "config.toml", "Path to the config file")
}
