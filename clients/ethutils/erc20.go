package ethutils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vegaprotocol/vega-monitoring/clients/ethutils/IERC20"
)

type ERC20Token struct {
	*IERC20.IERC20
	HexAddress string
}

func (c *EthClient) GetAssetPoolBalanceForToken(hexTokenAddress string) (*big.Int, error) {
	tokenClient, err := c.GetERC20(hexTokenAddress)
	if err != nil {
		return nil, err
	}
	return tokenClient.BalanceOf(c.ethConfig.AssetPoolAddress)
}

func (c *EthClient) GetERC20(hexTokenAddress string) (*ERC20Token, error) {
	tokenAddress := common.HexToAddress(hexTokenAddress)
	tokenClient, err := IERC20.NewIERC20(tokenAddress, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ERC20 token %s, %w", hexTokenAddress, err)
	}
	return &ERC20Token{
		IERC20:     tokenClient,
		HexAddress: hexTokenAddress,
	}, nil
}

func (t *ERC20Token) BalanceOf(hexAddress string) (*big.Int, error) {
	address := common.HexToAddress(hexAddress)
	balance, err := t.IERC20.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for %s for Token %s, %w", hexAddress, t.HexAddress, err)
	}
	return balance, nil
}
