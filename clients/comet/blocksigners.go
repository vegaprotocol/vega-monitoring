package comet

import (
	"fmt"
)

func (c *CometClient) GetBlockSigners(block int64) (*BlockSignersData, error) {
	response, err := c.requestCometCommit(block)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data for block %d, %w", block, err)
	}
	blockData, err := NewBlockSignersData(response)
	if err != nil {
		return nil, fmt.Errorf("failed to get block signers data for block %d, %w", block, err)
	}
	return &blockData, nil
}

func (c *CometClient) GetBlockSignersRange(fromBlock int64, toBlock int64) ([]BlockSignersData, error) {

	responses, err := c.requestCometCommitRange(fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data for blocks from %d to %d, %w", fromBlock, toBlock, err)
	}
	result := make([]BlockSignersData, len(responses))
	for _, response := range responses {
		blockData, err := NewBlockSignersData(response)
		if err != nil {
			return nil, fmt.Errorf("failed to get block signers data for blocks from %d to %d, %w", fromBlock, toBlock, err)
		}
		result[blockData.Height-fromBlock] = blockData
	}
	return result, nil
}
