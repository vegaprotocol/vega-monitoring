package types

import (
	"fmt"
)

type ETHNetwork string

const (
	ETHMainnet ETHNetwork = "mainnet"
	ETHSepolia ETHNetwork = "sepolia"
	ETHGoerli  ETHNetwork = "goerli"
	ETHRopsten ETHNetwork = "ropsten"
	ETHLocal   ETHNetwork = "local"
)

func (n ETHNetwork) IsValid() error {
	switch n {
	case ETHMainnet, ETHSepolia, ETHGoerli, ETHRopsten, ETHLocal:
		return nil
	}
	return fmt.Errorf("Invalid Ethereum network %s", n)
}

func GetEthNetworkForId(chainId string) (ETHNetwork, error) {
	switch chainId {
	case "1":
		return ETHMainnet, nil
	case "3":
		return ETHRopsten, nil
	case "5":
		return ETHGoerli, nil
	case "11155111":
		return ETHSepolia, nil
	}
	return "", fmt.Errorf("unknown Ethereum chain id: %s", chainId)
}
