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
	var ethNodes []config.EthereumNodeConfig
	cfg, log, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if args.EthNodeURLs != nil && len(args.EthNodeURLs) > 0 {
		for _, node := range args.EthNodeURLs {
			ethNodes = append(ethNodes, config.EthereumNodeConfig{RPCEndpoint: node})
		}
	} else if err != nil {
		return fmt.Errorf("Required --url flag or config.toml file")
	} else {
		ethNodes = cfg.Monitoring.EthereumNode
	}

	results := ethutils.CheckETHEndpointList(context.Background(), log, ethNodes)

	log.Info("Results:")
	for endpoint, isHealthy := range results {
		log.Info("-", zap.String("endpoint", endpoint), zap.Bool("is healthy", isHealthy))
	}

	return nil
}
