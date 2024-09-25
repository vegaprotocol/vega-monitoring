package update

import (
	"context"
	"fmt"
	"sync/atomic"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
	"go.uber.org/zap"
)

// TODO(fixme): We should not use global variables
var (
	earliestBlock atomic.Int64
)

func (us *UpdateService) UpdateCometTxsAllNew(ctx context.Context) error {
	return us.UpdateCometTxs(ctx, 0, 0)
}

func (us *UpdateService) UpdateCometTxs(ctx context.Context, fromBlock int64, toBlock int64) error {
	var err error
	serviceStore := us.storeService.NewCometTxs()
	logger := us.log.With(zap.String(UpdaterType, "UpdateCometTxs"))

	logger.Debug("getting network toBlock network height")
	blockStore := us.storeService.NewBlocks()

	latestBlockHeightForDataNode, err := blockStore.GetLastBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height for data node: %w", err)
	}

	latestBlockHeightForTendermint, err := us.readService.GetNetworkLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block for tendermint: %w", err)
	}

	// Make sure data node has latest block
	if toBlock <= 0 || toBlock > *latestBlockHeightForDataNode {
		toBlock = *latestBlockHeightForDataNode
	}
	if toBlock > latestBlockHeightForTendermint {
		toBlock = latestBlockHeightForTendermint
	}

	// get First Block
	if fromBlock <= 0 {
		logger.Debug("getting network fromBlock network height")
		lastProcessedBlock, err := serviceStore.GetLatestBlockInStore(context.Background())
		if err != nil {
			return fmt.Errorf("failed to Update Comet Txs, %w", err)
		}
		if lastProcessedBlock > 0 {
			fromBlock = lastProcessedBlock + 1
		} else {
			// No blocks in database - Get the first block from the Tendermint API
			earliestBlockForTendermint, err := us.readService.GetEarliestBlockHeight(ctx)
			if err != nil {
				return fmt.Errorf("failed to get the earliest comet block: %w", err)
			}

			fromBlock = toBlock - (BLOCK_NUM_IN_24h * 3)
			if fromBlock < earliestBlockForTendermint {
				fromBlock = earliestBlockForTendermint
			}
			if fromBlock <= 0 {
				fromBlock = 1
			}
		}
	}

	if earliestBlock.Load() == 0 {
		earliestBlockHeight, err := findEarliestABCIResponse(ctx, fromBlock, toBlock, us.readService)
		if err != nil {
			return fmt.Errorf("failed to get earliest blocks: %w", err)
		}
		earliestBlock.Store(earliestBlockHeight)
	}

	if fromBlock < earliestBlock.Load() {
		fromBlock = earliestBlock.Load()
	}

	if fromBlock > toBlock {
		return fmt.Errorf("cannot update Comet Txs, from block '%d' is greater than to block '%d'", fromBlock, toBlock)
	}

	// Update in batches
	logger.Debug(
		"Update Comet Txs in batches",
		zap.Int64("first block", fromBlock),
		zap.Int64("last block", toBlock),
	)

	var totalCount int = 0

	logger.Debugf("getting batch blocks from %d to %d", fromBlock, toBlock)
	for batchFirstBlock := fromBlock; batchFirstBlock <= toBlock; batchFirstBlock += 200 {
		batchLastBlock := batchFirstBlock + 199 // endBlock inclusive
		if batchLastBlock > toBlock {
			batchLastBlock = toBlock
		}
		count, err := UpdateCometTxsRange(ctx, batchFirstBlock, batchLastBlock, us.readService, serviceStore, logger)
		if err != nil {
			return fmt.Errorf("failed to update comet txs range: %w", err)
		}
		totalCount += count
	}

	logger.Debug(
		"Finished",
		zap.Int64("processed blocks", toBlock-fromBlock+1),
		zap.Int("total row count sotred in SQLStore", totalCount),
	)

	return nil
}

func findFirstExistingBlock(rangeStart int64, rangeEnd int64, blockExist func(blockNum int64) bool) int64 {
	end := rangeEnd

	// Block does not exist in given range
	if !blockExist(rangeEnd) {
		return -2
	}

	for start := rangeStart; start <= end; {
		if blockExist(start) && (start <= 1 || start == rangeStart || !blockExist(start-1)) {
			return start
		}
		middle := start + (end-start)/2

		if !blockExist(middle) {
			start = middle + 1
		} else {
			end = middle
		}
	}

	return -1
}

// There is situation when you started node with tendermint setting `discard_abci_responses = true` and then enabled it at some point
// then tendermint does not have information about the transactions. We use the `/block_results?height=` endpoint of the tendermint API,
// and for discarded abci responses this endpoint returns "could not find results for height #XYZ" error.
func findEarliestABCIResponse(ctx context.Context, rangeStart, rangeEnd int64, readService *read.ReadService) (int64, error) {
	firstExistingBlock := findFirstExistingBlock(rangeStart, rangeEnd, func(blockNum int64) bool {
		_, err := readService.GetTxsFromBlock(ctx, blockNum)

		return err == nil
	})
	if firstExistingBlock > 0 {
		return firstExistingBlock, nil
	}

	return -1, fmt.Errorf("there is no ABCI response in the comet storage")
}

func UpdateCometTxsRange(
	ctx context.Context,
	fromBlock int64,
	toBlock int64,
	readService *read.ReadService,
	serviceStore *sqlstore.CometTxs,
	logger *logging.Logger,
) (int, error) {
	txs, err := readService.GetCometTxs(ctx, fromBlock, toBlock)
	if err != nil {
		return -1, err
	}

	logger.Debug(
		"fetched data from CometBFT",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("tx count", len(txs)),
	)

	for _, tx := range txs {
		serviceStore.AddWithoutTime(tx)
	}

	storedData, err := serviceStore.FlushUpsertWithoutTime(context.Background())
	storedCount := len(storedData)
	if err != nil {
		return storedCount, fmt.Errorf("failed to flush comet txs range: %w", err)
	}
	logger.Debug(
		"stored data in SQLStore",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("row count", storedCount),
	)

	return storedCount, nil
}
