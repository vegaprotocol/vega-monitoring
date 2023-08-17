package ethutils

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils"
	"github.com/vegaprotocol/vega-monitoring/config"
	"go.uber.org/zap"
)

type CheckEthNodeArgs struct {
	*EthUtilsArgs
	EthNodeURLs []string
}

var checkEthNodeArgs CheckEthNodeArgs

// checkEthNodeCmd represents the checkEthNode command
var checkEthNodeCmd = &cobra.Command{
	Use:   "check-eth-node",
	Short: "Check if Ethereum Node is healthy",
	Long:  `Check if Ethereum Node is healthy`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunCheckEthNode(checkEthNodeArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	EthUtilsCmd.AddCommand(checkEthNodeCmd)
	checkEthNodeArgs.EthUtilsArgs = &ethUtilsArgs

	checkEthNodeCmd.PersistentFlags().StringSliceVar(&checkEthNodeArgs.EthNodeURLs, "url", nil, "Comma separated list of Ethereum RPC endpoints. If empty: config file will be used instead")
}

func RunCheckEthNode(args CheckEthNodeArgs) error {
	ethNodes := args.EthNodeURLs
	cfg, log, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if ethNodes == nil {
		if err != nil {
			return fmt.Errorf("Required --url flag or config.toml file")
		}
		ethNodes = cfg.Monitoring.EthereumNodes
	}

	results := ethutils.CheckETHEndpointList(context.Background(), log, ethNodes)

	log.Info("Results:")
	for endpoint, isHealthy := range results {
		log.Info("-", zap.String("endpoint", endpoint), zap.Bool("is healthy", isHealthy))
	}

	return nil
}
