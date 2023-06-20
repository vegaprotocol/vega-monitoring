package ethutils

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vegaprotocol/vega-monitoring/config"
)

type EthClient struct {
	ethConfig *config.EthereumConfig
	log       *logging.Logger
	client    *ethclient.Client
}

func NewEthClient(ethConfig *config.EthereumConfig, log *logging.Logger) (*EthClient, error) {
	client, err := ethclient.DialContext(context.Background(), ethConfig.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum Node, %w", err)
	}
	return &EthClient{
		ethConfig: ethConfig,
		log:       log,
		client:    client,
	}, nil
}
