package service

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type ServiceArgs struct {
	*rootCmd.RootArgs
}

var serviceArgs ServiceArgs

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service to gather data and put it into the database",
	Long:  `Service to gather data and put it into the database`,
}

func init() {
	serviceArgs.RootArgs = &rootCmd.Args
}
