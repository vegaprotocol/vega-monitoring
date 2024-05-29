package read

import "math/big"

func (s *ReadService) GetAssetPoolBalanceForTokenFromEthereum(tokenAddress, assetPoolAddress string) (*big.Int, error) {
	return s.ethClient.GetAssetPoolBalanceForToken(tokenAddress, assetPoolAddress)
}

func (s *ReadService) GetAssetPoolBalanceForTokenFromArbitrum(tokenAddress, assetPoolAddress string) (*big.Int, error) {
	return s.arbitrumClient.GetAssetPoolBalanceForToken(tokenAddress, assetPoolAddress)
}
