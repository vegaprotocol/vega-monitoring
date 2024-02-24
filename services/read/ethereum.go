package read

import "math/big"

func (s *ReadService) GetAssetPoolBalanceForToken(tokenAddress, assetPoolAddress string) (*big.Int, error) {
	return s.ethClient.GetAssetPoolBalanceForToken(tokenAddress, assetPoolAddress)
}
