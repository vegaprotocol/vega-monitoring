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
	abiJSON               string
	lastCalledBlock       *big.Int
	abiObject             abi.ABI
	initialCallPastBlocks uint64
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
		abiJSON:               abiJSON,
		initialCallPastBlocks: initialCallPastBlocks,
		abiObject:             contractAbi,
		lastCalledBlock:       nil,
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

func (e *EventsCount) CountEvents(ctx context.Context, client *EthClient) (map[string]uint64, error) {
	result := map[string]uint64{
		AllEvents: 0,
	}

	height, err := client.Height(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ethereum height for initial call for the %s smart contract: %w", e.contractAddressString, err)
	}

	heightBigInt := big.NewInt(0).SetUint64(height)

	if e.lastCalledBlock == nil {
		e.lastCalledBlock = big.NewInt(0).Sub(heightBigInt, big.NewInt(0).SetUint64(e.initialCallPastBlocks))
	}

	if big.NewInt(0).Sub(heightBigInt, e.lastCalledBlock).Cmp(big.NewInt(1)) < 0 {
		// Ethereum did not make any block
		return result, nil
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{e.contractAddress},
		FromBlock: e.lastCalledBlock,
		ToBlock:   heightBigInt,
	}

	logs, err := client.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter ethereum logs for contract %s: %w", e.contractAddressString, err)
	}

	for _, vLog := range logs {
		result[AllEvents] = result[AllEvents] + 1

		// https://docs.alchemy.com/docs/deep-dive-into-eth_getlogs
		event, err := e.abiObject.EventByID(vLog.Topics[0])
		if err == nil {
			// event can be deducted from the ABI
			eventName := event.Name
			val, ok := result[eventName]
			if !ok {
				val = 0
			}
			result[eventName] = val + 1

			continue
		}
	}

	return result, nil
}
