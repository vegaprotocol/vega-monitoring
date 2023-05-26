package ethutils

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	ethURL string
	client ethclient.Client
}

func NewEthClient(ethURL string) (*EthClient, error) {
	client, err := ethclient.DialContext(context.Background(), ethURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum Node, %w", err)
	}
	return &EthClient{
		ethURL: ethURL,
		client: *client,
	}, nil
}
