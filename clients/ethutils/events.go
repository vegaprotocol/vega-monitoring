package ethutils

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vegaprotocol/vega-monitoring/config"
)

const (
	AllEvents                    = "*"
	DefaultInitialCallPastBlocks = 32 // call X past blocks
)

type EventsCount struct {
	name string

	contractAddressString string
	contractAddress       common.Address

	abiJSON   string
	abiObject abi.ABI

	lastCalledBlock       *big.Int
	initialCallPastBlocks uint64

	result map[string]uint64
}

func NewEventsCount(name string, address string, abiJSON string, initialCallPastBlocks uint64) (*EventsCount, error) {
	contractAbi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI object from JSON: %w", err)
	}

	if initialCallPastBlocks < 0 {
		return nil, fmt.Errorf("initialCallPastBlocks cannot be smaller than 0")
	}

	return &EventsCount{
		name: name,

		contractAddressString: address,
		contractAddress:       common.HexToAddress(address),

		abiJSON:   abiJSON,
		abiObject: contractAbi,

		initialCallPastBlocks: initialCallPastBlocks,
		lastCalledBlock:       nil,

		result: map[string]uint64{
			AllEvents: 0,
		},
	}, nil
}

func NewEventsCountFromConfig(cfg config.EthEvents) (*EventsCount, error) {
	return NewEventsCount(
		cfg.Name,
		cfg.ContractAddress,
		cfg.ABI,
		DefaultInitialCallPastBlocks,
	)
}

func (e *EventsCount) CallFilterLogs(ctx context.Context, client *EthClient) error {
	height, err := client.Height(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ethereum height for initial call for the %s smart contract: %w", e.contractAddressString, err)
	}

	heightBigInt := big.NewInt(0).SetUint64(height)

	if e.lastCalledBlock == nil {
		e.lastCalledBlock = big.NewInt(0).Sub(heightBigInt, big.NewInt(0).SetUint64(e.initialCallPastBlocks))
	}

	if big.NewInt(0).Sub(heightBigInt, e.lastCalledBlock).Cmp(big.NewInt(1)) < 0 {
		// Ethereum did not make any block
		return nil
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{e.contractAddress},
		FromBlock: e.lastCalledBlock,
		ToBlock:   heightBigInt,
	}

	logs, err := client.client.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter ethereum logs for contract %s: %w", e.contractAddressString, err)
	}

	for _, vLog := range logs {
		e.result[AllEvents] = e.result[AllEvents] + 1

		// Event is not indexed
		if len(vLog.Topics) < 1 {
			continue
		}

		// https://docs.alchemy.com/docs/deep-dive-into-eth_getlogs
		event, err := e.abiObject.EventByID(vLog.Topics[0])
		if err == nil {
			// event can be deducted from the ABI
			eventName := event.Name
			val, ok := e.result[eventName]
			if !ok {
				val = 0
			}
			e.result[eventName] = val + 1

			continue
		}
	}

	return nil
}

func (e EventsCount) Count() map[string]uint64 {
	return e.result
}
