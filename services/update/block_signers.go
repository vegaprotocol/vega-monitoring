package update

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"github.com/vegaprotocol/vega-monitoring/sqlstore"
	"go.uber.org/zap"
)

const (
	BLOCK_NUM_IN_24h int64 = 123500 // 24[h/day] * 60[min/h] * 60[sec/h] / 0.7[block/sec]
)

func (us *UpdateService) UpdateBlockSignersAllNew(ctx context.Context) error {
	return us.UpdateBlockSigners(ctx, 0, 0)
}

func (us *UpdateService) UpdateBlockSigners(ctx context.Context, fromBlock int64, toBlock int64) error {
	var err error
	blockSigner := us.storeService.NewBlockSigner()

	logger := us.log.With(zap.String(UpdaterType, "UpdateBlockSigners"))

	// get Last Block
	if toBlock <= 0 {
		logger.Debug("getting network toBlock network height")
		toBlock, err = us.readService.GetNetworkLatestBlockHeight()
		if err != nil {
			return fmt.Errorf("failed to Update Block Signers, %w", err)
		}
	}

	// get First Block
	if fromBlock <= 0 {
		logger.Debug("getting network fromBlock network height")
		lastProcessedBlock, err := blockSigner.GetLastestBlockInStore(context.Background())
		if err != nil {
			return fmt.Errorf("failed to Update Block Signers, %w", err)
		}

		// We have some processed blocks in the data-base
		if lastProcessedBlock > 0 {
			fromBlock = lastProcessedBlock + 1
		} else {
			// No blocks in database - Get the first block from the Tendermint API
			earliestBlock, err := us.readService.GetEarliestBlockHeight(ctx)
			if err != nil {
				return fmt.Errorf("failed to get the earliest comet block: %w", err)
			}

			fromBlock = toBlock - (BLOCK_NUM_IN_24h * 3)
			if fromBlock < earliestBlock {
				fromBlock = earliestBlock
			}
			if fromBlock <= 0 {
				fromBlock = 1
			}

		}
	}
	if fromBlock > toBlock {
		return fmt.Errorf("cannot update Block Signers, from block '%d' is greater than to block '%d'", fromBlock, toBlock)
	}
	// Update in batches
	logger.Info(
		"Update Block Signers in batches",
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
		count, err := UpdateBlockRange(batchFirstBlock, batchLastBlock, us.readService, blockSigner, logger)
		if err != nil {
			return fmt.Errorf("failed to update block range: %w", err)
		}
		totalCount += count
	}

	logger.Info(
		"Finished",
		zap.Int64("processed blocks", toBlock-fromBlock+1),
		zap.Int("total row count sotred in SQLStore", totalCount),
	)

	return nil
}

func UpdateBlockRange(
	fromBlock int64,
	toBlock int64,
	readService *read.ReadService,
	blockSignerStore *sqlstore.BlockSigner,
	logger *logging.Logger,
) (int, error) {

	blocks, err := readService.GetBlockSigners(fromBlock, toBlock)
	if err != nil {
		return -1, fmt.Errorf("failed to get block signers: %w", err)
	}
	logger.Info(
		"fetched data from CometBFT",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("block count", len(blocks)),
	)

	for _, block := range blocks {
		valData, err := readService.GetValidatorForAddressAtBlock(block.ProposerAddress, block.Height)
		if err != nil {
			return 0, fmt.Errorf("failed to get validator for address at block(1): %w", err)
		}
		blockSignerStore.Add(&entities.BlockSigner{
			VegaTime: block.Time,
			Role:     entities.BlockSignerRoleProposer,
			TmPubKey: valData.TmPubKey,
		})
		for _, signerAddress := range block.SignerAddresses {
			valData, err := readService.GetValidatorForAddressAtBlock(signerAddress, block.Height)
			if err != nil {
				return 0, fmt.Errorf("failed to get validator for address at block(2): %w", err)
			}
			blockSignerStore.Add(&entities.BlockSigner{
				VegaTime: block.Time,
				Role:     entities.BlockSignerRoleSigner,
				TmPubKey: valData.TmPubKey,
			})
		}
	}

	storedData, err := blockSignerStore.FlushUpsert(context.Background())
	storedCount := len(storedData)
	if err != nil {
		return storedCount, fmt.Errorf("failed to flush block signers: %w", err)
	}
	logger.Info(
		"stored data in SQLStore",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("row count", storedCount),
	)

	return storedCount, nil
}
