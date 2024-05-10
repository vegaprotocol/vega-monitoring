package comet

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"
)

//
//
//

var (
	ErrMissingSubmitter = errors.New("missing submitter in the comet block")
	ErrMissingCommand   = errors.New("missing command in the comet block")
)

type CometTx struct {
	Code       int
	Info       *string
	Submitter  string
	Command    string
	Attributes map[string]string
	Height     int64
	HeightIdx  int
}

func (c *CometClient) GetLastBlockTxs(ctx context.Context) ([]CometTx, error) {
	return c.GetTxsForBlock(ctx, 0)
}

func (c *CometClient) GetTxsForBlock(ctx context.Context, block int64) ([]CometTx, error) {
	txList, err := c.GetTxsForBlockNotFiltered(ctx, block)
	if err != nil {
		return nil, err
	}
	return RemoveExcludedTxTypes(txList), nil
}

func (c *CometClient) GetTxsForBlockRange(ctx context.Context, fromBlock int64, toBlock int64) ([]CometTx, error) {
	txList, err := c.GetTxsForBlockRangeNotFiltered(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}
	return RemoveExcludedTxTypes(txList), nil
}

var (
	EXCLUDE_COMMAND_TYPES = []string{
		"Submit Order",
		"Cancel Order",
		"Amend Order",
		"Withdraw",
		"Proposal",
		"Vote on Proposal",
		// "Register new Node",
		// "Node Vote",
		"Node Signature",
		"Liquidity Provision Order",
		"Cancel LiquidityProvision Order",
		"Amend LiquidityProvision Order",
		// "Chain Event",
		"Submit Oracle Data",
		"Delegate",
		"Undelegate",
		"Key Rotate Submission",
		// "State Variable Proposal",
		"Transfer Funds",
		"Cancel Transfer Funds",
		// "Validator Heartbeat",
		"Ethereum Key Rotate Submission",
		"Protocol Upgrade",
		// "Issue Signatures",
		// "Batch Market Instructions",
	}
)

func RemoveExcludedTxTypes(txs []CometTx) []CometTx {
	result := []CometTx{}
	for _, tx := range txs {
		if slices.Contains(EXCLUDE_COMMAND_TYPES, tx.Command) {
			continue
		}
		result = append(result, tx)
	}
	return result
}

func (c *CometClient) GetTxsForBlockNotFiltered(ctx context.Context, block int64) ([]CometTx, error) {
	response, err := c.requestBlockResults(ctx, block)
	if err != nil {
		return nil, err
	}
	txsList, err := parseBlockResultsResponse(response)
	if err != nil {
		return nil, err
	}
	return txsList, nil
}

func (c *CometClient) GetTxsForBlockRangeNotFiltered(ctx context.Context, fromBlock int64, toBlock int64) ([]CometTx, error) {
	responses, err := c.requestBlockResultsRange(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get block result data for blocks from %d to %d, %w", fromBlock, toBlock, err)
	}
	result := []CometTx{}

	for _, response := range responses {
		txsList, err := parseBlockResultsResponse(response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block result response for block %s: %w", response.Result.Height, err)
		}
		result = append(result, txsList...)
	}
	return result, nil
}

//
// Helper functions
//

func parseBlockResultsResponse(response blockResultsResponse) (result []CometTx, err error) {
	height, err := strconv.ParseInt(response.Result.Height, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse Height '%s' to int, from: %+v.", response.Result.Height, response)
		return
	}
	result = make([]CometTx, len(response.Result.TxsResults))
	for i, tx := range response.Result.TxsResults {

		result[i].Height = height
		result[i].HeightIdx = i
		result[i].Code = tx.Code
		if len(tx.Info) > 0 {
			info := tx.Info
			result[i].Info = &info
		}

		submitter, command, attrs, err := parseBlockResultsResponseTxEvent(tx.Events)
		if err != nil {
			if errors.Is(err, ErrMissingSubmitter) {
				// Block may not contain transaction with submitter
				continue
			}
			return nil, fmt.Errorf("failed to parse Tx Events for block %d, %w", height, err)
		}

		result[i].Submitter = submitter
		result[i].Command = command
		result[i].Attributes = attrs
	}
	return
}

func parseBlockResultsResponseTxEvent(
	events []blockResultsResponseTxEvent,
) (submitter string, command string, attributes map[string]string, err error) {

	attributes = map[string]string{}

	for _, event := range events {
		for _, attr := range event.Attributes {
			if event.Type == "tx" && attr.Key == "submitter" {
				submitter = attr.Value
			} else if event.Type == "command" && attr.Key == "type" {
				command = attr.Value
			} else {
				attributes[attr.Key] = attr.Value
			}
		}
	}

	if len(attributes) == 0 {
		attributes = nil
	}
	if len(submitter) == 0 {
		err = fmt.Errorf("%w: %+v", ErrMissingSubmitter, events)
		return
	}
	if len(command) == 0 {
		err = fmt.Errorf("%w: %+v", ErrMissingCommand, events)
		return
	}
	return
}
