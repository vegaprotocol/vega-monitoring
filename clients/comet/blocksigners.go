package comet

import (
	"context"
	"fmt"
)

func (c *CometClient) GetLatestBlockSigners(ctx context.Context) (*BlockSignersData, error) {
	return c.GetBlockSigners(ctx, -1)
}

func (c *CometClient) GetBlockSigners(ctx context.Context, block int64) (*BlockSignersData, error) {
	response, err := c.requestCommit(ctx, block)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data for block %d, %w", block, err)
	}
	blockData, err := newBlockSignersData(response)
	if err != nil {
		return nil, fmt.Errorf("failed to get block signers data for block %d, %w", block, err)
	}
	return &blockData, nil
}

func (c *CometClient) GetBlockSignersRange(ctx context.Context, fromBlock int64, toBlock int64) ([]BlockSignersData, error) {

	responses, err := c.requestCommitRange(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data for blocks from %d to %d, %w", fromBlock, toBlock, err)
	}
	result := make([]BlockSignersData, len(responses))
	for _, response := range responses {
		blockData, err := newBlockSignersData(response)
		if err != nil {
			return nil, fmt.Errorf("failed to get block signers data for blocks from %d to %d, %w", fromBlock, toBlock, err)
		}
		result[blockData.Height-fromBlock] = blockData
	}
	return result, nil
}
