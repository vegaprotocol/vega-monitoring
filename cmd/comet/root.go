package comet

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type CometArgs struct {
	*rootCmd.RootArgs
	ApiURL string
}

var cometArgs CometArgs

var CometCmd = &cobra.Command{
	Use:   "comet",
	Short: "Interact with CometBFT",
	Long:  `Interact with CometBFT`,
}

func init() {
	cometArgs.RootArgs = &rootCmd.Args
	CometCmd.PersistentFlags().StringVar(&cometArgs.ApiURL, "api-url", "", "CometBFT API URL")
}
