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
	DefaultInitialCallPastBlocks = 128
	DefaultMaxBlocks             = 9999
)

type EventsCounter struct {
	name string

	contractAddressString string
	contractAddress       common.Address

	abiJSON   string
	abiObject abi.ABI

	lastCalledBlock       *big.Int
	initialCallPastBlocks uint64
	maxBlocks             uint64

	// Result keeps information about all seen events since monitoring started
	result map[string]uint64
}

func NewEventsCounter(name string, address string, abiJSON string, initialCallPastBlocks uint64, maxBlocks uint64) (*EventsCounter, error) {
	contractAbi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI object from JSON: %w", err)
	}

	return &EventsCounter{
		name: name,

		contractAddressString: address,
		contractAddress:       common.HexToAddress(address),

		abiJSON:   abiJSON,
		abiObject: contractAbi,

		initialCallPastBlocks: initialCallPastBlocks,
		maxBlocks:             maxBlocks,
		lastCalledBlock:       nil,

		result: map[string]uint64{
			AllEvents: 0,
		},
	}, nil
}

func NewEventsCounterFromConfig(cfg config.EthEvents) (*EventsCounter, error) {
	pastBlocks := cfg.InitialBlocksToScan
	if pastBlocks < 1 {
		pastBlocks = DefaultInitialCallPastBlocks
	}

	maxBlocks := cfg.MaxBlocksToFilter
	if maxBlocks < 1 {
		maxBlocks = DefaultMaxBlocks
	}

	return NewEventsCounter(
		cfg.Name,
		cfg.ContractAddress,
		cfg.ABI,
		pastBlocks,
		uint64(maxBlocks),
	)
}

func (e *EventsCounter) CallFilterLogs(ctx context.Context, client *EthClient) error {
	currentHeight, err := client.Height(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ethereum height for initial call for the %s smart contract: %w", e.contractAddressString, err)
	}

	currentEvmBlockHeight := big.NewInt(0).SetUint64(currentHeight)

	if e.lastCalledBlock == nil {
		e.lastCalledBlock = big.NewInt(0).Sub(currentEvmBlockHeight, big.NewInt(0).SetUint64(e.initialCallPastBlocks))
	}

	if big.NewInt(0).Sub(currentEvmBlockHeight, e.lastCalledBlock).Cmp(big.NewInt(1)) < 0 {
		// Ethereum did not make any block
		return nil
	}

	toBlock := big.NewInt(0).Set(currentEvmBlockHeight)
	// Lets call max 9999 blocks as some RPC providers limit filter to 10k blocks
	if big.NewInt(0).Sub(toBlock, e.lastCalledBlock).Cmp(big.NewInt(int64(e.maxBlocks))) > 0 {
		toBlock = big.NewInt(0).Add(e.lastCalledBlock, big.NewInt(int64(e.maxBlocks)))
	}

	// toBlock is somehow higher than current ethereum block
	if toBlock.Cmp(currentEvmBlockHeight) > 0 {
		toBlock = big.NewInt(0).Set(currentEvmBlockHeight)
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{e.contractAddress},
		FromBlock: e.lastCalledBlock,
		ToBlock:   toBlock,
	}

	logs, err := client.client.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf(
			"failed to filter ethereum logs for contract %s for block <%s; %s>: %s",
			e.contractAddressString,
			e.lastCalledBlock.String(),
			toBlock.String(),
			err.Error(),
		)
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
	e.lastCalledBlock = big.NewInt(0).Set(toBlock)

	return nil
}

// Count returns result
func (e EventsCounter) Count() map[string]uint64 {
	return e.result
}

func (e EventsCounter) Name() string {
	return e.name
}

func (e EventsCounter) ContractAddress() string {
	return e.contractAddressString
}
