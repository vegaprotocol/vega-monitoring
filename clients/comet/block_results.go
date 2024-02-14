package comet

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"
)

//
//
//

type CometTx struct {
	Code       int
	Info       *string
	Submitter  string
	Command    string
	Attributes map[string]string
	Height     int64
	HeightIdx  int
}

func (c *CometClient) GetLastBlockTxs() ([]CometTx, error) {
	return c.GetTxsForBlock(0)
}

func (c *CometClient) GetTxsForBlock(block int64) ([]CometTx, error) {
	txList, err := c.GetTxsForBlockNotFiltered(block)
	if err != nil {
		return nil, err
	}
	return RemoveExcludedTxTypes(txList), nil
}

func (c *CometClient) GetTxsForBlockRange(fromBlock int64, toBlock int64) ([]CometTx, error) {
	txList, err := c.GetTxsForBlockRangeNotFiltered(fromBlock, toBlock)
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

func (c *CometClient) GetTxsForBlockNotFiltered(block int64) ([]CometTx, error) {
	response, err := c.requestBlockResults(block)
	if err != nil {
		return nil, err
	}
	txsList, err := parseBlockResultsResponse(response)
	if err != nil {
		return nil, err
	}
	return txsList, nil
}

func (c *CometClient) GetTxsForBlockRangeNotFiltered(fromBlock int64, toBlock int64) ([]CometTx, error) {
	responses, err := c.requestBlockResultsRange(fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get block result data for blocks from %d to %d, %w", fromBlock, toBlock, err)
	}
	result := []CometTx{}

	for _, response := range responses {
		txsList, err := parseBlockResultsResponse(response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block result response for block %d: %w", response.Result.Height, err)
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

		result[i].Submitter, result[i].Command, result[i].Attributes, err = parseBlockResultsResponseTxEvent(tx.Events)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Tx Events for block %d, %w", height, err)
		}
	}
	return
}

func parseBlockResultsResponseTxEvent(
	events []blockResultsResponseTxEvent,
) (submitter string, command string, attributes map[string]string, err error) {

	attributes = map[string]string{}

	for _, event := range events {
		for _, attr := range event.Attributes {
			var bKey, bValue []byte
			bKey, err = base64.StdEncoding.DecodeString(attr.Key)
			if err != nil {
				err = fmt.Errorf("failed to decode base64 attribute key %s, %w", attr.Key, err)
				return
			}
			key := string(bKey[:])
			bValue, err = base64.StdEncoding.DecodeString(attr.Value)
			if err != nil {
				err = fmt.Errorf("failed to decode base64 attribute value %s, %w", attr.Value, err)
				return
			}
			value := string(bValue[:])
			if event.Type == "tx" && key == "submitter" {
				submitter = value
			} else if event.Type == "command" && key == "type" {
				command = value
			} else {
				attributes[key] = value
			}
		}
	}

	if len(attributes) == 0 {
		attributes = nil
	}
	if len(submitter) == 0 {
		err = fmt.Errorf("missing submitter in Comet Block Result tx: %+v", events)
		return
	}
	if len(command) == 0 {
		err = fmt.Errorf("missing command in Comet Block Result tx: %+v", events)
		return
	}
	return
}
