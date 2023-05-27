package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RootArgs struct {
	Debug bool
}

var Args RootArgs

var RootCmd = &cobra.Command{
	Use:   "data-metrics-store",
	Short: "",
	Long:  ``,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run data-metrics-store '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.PersistentFlags().BoolVar(&Args.Debug, "debug", false, "Print debug logs")
}
