package healthcheck

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type HealthcheckArgs struct {
	*rootCmd.RootArgs
}

var healthcheckArgs HealthcheckArgs

var HealthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Healthcheck to gather info about local node and expose it trough http and metrics",
	Long:  `Healthcheck to gather info about local node and expose it trough http and metrics`,
}

func init() {
	healthcheckArgs.RootArgs = &rootCmd.Args
}
