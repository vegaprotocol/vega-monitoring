package ethutils

import (
	"context"
	"fmt"
	"math/big"

	"code.vegaprotocol.io/vega/logging"
	"github.com/ethereum/go-ethereum/common"
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

func (c *EthClient) BalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	return c.client.BalanceAt(ctx, account, nil)
}

func (c *EthClient) BalanceWithoutZerosAt(ctx context.Context, account common.Address) (float64, error) {
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance without zeros at for account %s: %w", account.String(), err)
	}

	result := big.NewInt(0).Div(
		balance, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(12), nil))

	return float64(result.Int64()) / 1000000, nil
}

func (c *EthClient) Call(ctx context.Context, call *EthCall) (interface{}, error) {
	return call.Call(ctx, c.client)
}

func (c *EthClient) Height(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *EthClient) Ready(ctx context.Context) (bool, error) {
	syncProcess, err := c.client.SyncProgress(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if node is syncing: %w", err)
	}

	if syncProcess != nil {
		return false, fmt.Errorf("node is still syncing")
	}

	return true, nil
}
