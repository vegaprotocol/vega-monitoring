package comet

import (
	"log"

	"github.com/spf13/cobra"
)

type CometArgs struct {
	ApiURL string
}

var cometArgs CometArgs

var CometCmd = &cobra.Command{
	Use:   "comet",
	Short: "Interact with CometBFT",
	Long:  `Interact with CometBFT`,
}

func init() {
	CometCmd.PersistentFlags().StringVar(&cometArgs.ApiURL, "api-url", "", "CometBFT API URL")
	if err := CometCmd.MarkPersistentFlagRequired("api-url"); err != nil {
		log.Fatalf("%v\n", err)
	}
}
