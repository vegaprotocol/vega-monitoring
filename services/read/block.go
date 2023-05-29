package read

import (
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
)

func (s *ReadService) GetNetworkLatestBlockHeight() (int64, error) {
	blockData, err := s.cometClient.GetLatestBlockSigners()
	if err != nil {
		return 0, err
	}
	return blockData.Height, nil
}

func (s *ReadService) GetBlockSigners(fromBlock int64, toBlock int64) ([]comet.BlockSignersData, error) {
	return s.cometClient.GetBlockSignersRange(fromBlock, toBlock)
}
