package etherscan

import (
	"fmt"
	"math/big"
)

type tokenBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=0x57d90b64a1a57749b0f932f1a3395792e12e7055&address=0xe04f27eb70e025b78871a2ad7eabe85e61212761&tag=latest

func (c *EtherscanClient) requestTokenBalance(address string, tokenAddress string) (responsePayload tokenBalanceResponse, err error) {

	err = c.sendRequest(map[string]string{
		"module":          "account",
		"action":          "tokenbalance",
		"address":         address,
		"contractaddress": tokenAddress,
		"tag":             "latest",
	}, &responsePayload)

	if responsePayload.Status != "1" {
		err = fmt.Errorf("failed Etherscan reqeust to get Token Balance '%s' for Address '%s'. Response status: %s. API response Error: %s", tokenAddress, address, responsePayload.Status, responsePayload.Message)
	}

	return
}

func (c *EtherscanClient) GetTokenBalance(address string, tokenAddress string) (*big.Int, error) {
	response, err := c.requestTokenBalance(address, tokenAddress)
	if err != nil {
		return nil, err
	}
	balance, ok := new(big.Int).SetString(response.Result, 0)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance '%s' to big.Int.", response.Result)
	}
	return balance, nil
}
