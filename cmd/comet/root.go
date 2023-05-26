package comet

import (
	"log"

	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
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
	if err := CometCmd.MarkPersistentFlagRequired("api-url"); err != nil {
		log.Fatalf("%v\n", err)
	}
}
