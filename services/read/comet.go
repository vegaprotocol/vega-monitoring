package read

import (
	"context"

	"github.com/vegaprotocol/vega-monitoring/clients/comet"
)

func (s *ReadService) GetNetworkLatestBlockHeight(ctx context.Context) (int64, error) {
	blockData, err := s.cometClient.GetLatestBlockSigners(ctx)
	if err != nil {
		return 0, err
	}
	return blockData.Height, nil
}

func (s *ReadService) GetBlockSigners(ctx context.Context, fromBlock int64, toBlock int64) ([]comet.BlockSignersData, error) {
	return s.cometClient.GetBlockSignersRange(ctx, fromBlock, toBlock)
}

func (s *ReadService) GetCometTxs(ctx context.Context, fromBlock int64, toBlock int64) ([]comet.CometTx, error) {
	return s.cometClient.GetTxsForBlockRange(ctx, fromBlock, toBlock)
}

func (s *ReadService) GetValidatorForAddressAtBlock(ctx context.Context, address string, block int64) (*comet.ValidatorData, error) {
	return s.cometClient.GetValidatorForAddressAtBlock(ctx, address, block)
}

func (s *ReadService) GetEarliestBlockHeight(ctx context.Context) (int64, error) {
	return s.cometClient.EarliestBlockHeight(ctx)
}

func (s *ReadService) GetLatestLocalBlockHeight(ctx context.Context) (int64, error) {
	return s.cometClient.LatestLocalBlockHeight(ctx)
}
