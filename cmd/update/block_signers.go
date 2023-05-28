package update

import (
	"context"
	"fmt"
	"log"
	"os"

	"code.vegaprotocol.io/vega/logging"
	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/comet"
	"github.com/vegaprotocol/data-metrics-store/config"
	"github.com/vegaprotocol/data-metrics-store/entities"
	"github.com/vegaprotocol/data-metrics-store/sqlstore"
	"go.uber.org/zap"
)

type BlockSignersArgs struct {
	*UpdateArgs
	FromBlock int64
	ToBlock   int64
}

var blockSignersArgs BlockSignersArgs

// getBlockCmd represents the getBlock command
var blockSignersCmd = &cobra.Command{
	Use:   "block-signers",
	Short: "Get data from CometBFT REST API and store it in SQLStore",
	Long:  `Get data from CometBFT REST API and store it in SQLStore`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunBlockSigners(blockSignersArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	UpdateCmd.AddCommand(blockSignersCmd)
	blockSignersArgs.UpdateArgs = &updateArgs

	blockSignersCmd.PersistentFlags().Int64Var(&blockSignersArgs.FromBlock, "from-block", 1, "First block to get")
	if err := blockSignersCmd.MarkPersistentFlagRequired("from-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
	blockSignersCmd.PersistentFlags().Int64Var(&blockSignersArgs.ToBlock, "to-block", 1, "Last block to get")
	if err := blockSignersCmd.MarkPersistentFlagRequired("to-block"); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func RunBlockSigners(args BlockSignersArgs) error {
	cfg, logger, err := config.GetConfigAndLogger(args.ConfigFilePath, args.Debug)
	if err != nil {
		return err
	}

	cometClient := comet.NewCometClient(cfg.CometBFT.ApiURL)

	connSource, err := sqlstore.NewTransactionalConnectionSource(logger, cfg.GetConnectionConfig())
	if err != nil {
		return err
	}
	blockSignerStore := sqlstore.NewBlockSigner(connSource)

	var totalCount int = 0

	for fromBlock := args.FromBlock; fromBlock <= args.ToBlock; fromBlock += 200 {
		toBlock := fromBlock + 199 // toBlock inclusive
		if args.ToBlock < toBlock {
			toBlock = args.ToBlock
		}
		count, err := UpdateBlockRange(fromBlock, toBlock, cometClient, blockSignerStore, logger)
		if err != nil {
			return err
		}
		totalCount += count
	}

	logger.Info(
		"Finished",
		zap.Int64("blocks", args.ToBlock-args.FromBlock+1),
		zap.Int("total row count sotred in SQLStore", totalCount),
	)

	return nil
}

func UpdateBlockRange(
	fromBlock int64,
	toBlock int64,
	cometClient *comet.CometClient,
	blockSignerStore *sqlstore.BlockSigner,
	logger *logging.Logger,
) (int, error) {

	blocks, err := cometClient.GetBlockSignersRange(fromBlock, toBlock)
	if err != nil {
		return -1, err
	}
	logger.Info(
		"fetched data from CometBFT",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("block count", len(blocks)),
	)

	for _, block := range blocks {
		blockSignerStore.Add(&entities.BlockSigner{
			VegaTime:  block.Time,
			Height:    block.Height,
			Role:      entities.BlockSignerRoleProposer,
			TmAddress: block.ProposerAddress,
			//TmPubKey: vega_entities.TendermintPublicKey(block.ProposerAddress),
		})
		for _, signerAddress := range block.SignerAddresses {
			blockSignerStore.Add(&entities.BlockSigner{
				VegaTime:  block.Time,
				Height:    block.Height,
				Role:      entities.BlockSignerRoleSigner,
				TmAddress: signerAddress,
				//TmPubKey: vega_entities.TendermintPublicKey(block.ProposerAddress),
			})
		}
	}

	storedData, err := blockSignerStore.FlushUpsert(context.Background())
	storedCount := len(storedData)
	if err != nil {
		return storedCount, err
	}
	logger.Info(
		"stored data in SQLStore",
		zap.Int64("from-block", fromBlock),
		zap.Int64("to-block", toBlock),
		zap.Int("row count", storedCount),
	)

	return storedCount, nil
}
