package read

import "math/big"

func (s *ReadService) GetAssetPoolBalanceForToken(tokenAddress string) (*big.Int, error) {
	return s.ethClient.GetAssetPoolBalanceForToken(tokenAddress)
}
