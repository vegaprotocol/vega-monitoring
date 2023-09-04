package grafana

import (
	"log"

	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type GrafanaArgs struct {
	*rootCmd.RootArgs
	ApiURL   string
	ApiToken string
}

var grafanaArgs GrafanaArgs

var GrafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Interact with Grafana",
	Long:  `Interact with Grafana`,
}

func init() {
	grafanaArgs.RootArgs = &rootCmd.Args
	GrafanaCmd.PersistentFlags().StringVar(&grafanaArgs.ApiURL, "url", "", "Grafana API URL")
	if err := GrafanaCmd.MarkPersistentFlagRequired("url"); err != nil {
		log.Fatalf("%v\n", err)
	}
	GrafanaCmd.PersistentFlags().StringVar(&grafanaArgs.ApiToken, "api-token", "", "Grafana Service Account access Token")
}
