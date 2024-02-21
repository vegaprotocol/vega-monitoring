package ethutils

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	log    *logging.Logger
	client *ethclient.Client
}

func NewEthClient(rpcAddress string, log *logging.Logger) (*EthClient, error) {
	client, err := ethclient.DialContext(context.Background(), rpcAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum Node, %w", err)
	}
	return &EthClient{
		log:    log,
		client: client,
	}, nil
}
