package update

import (
	"context"
	"fmt"

	"github.com/vegaprotocol/vega-monitoring/clients/datanode"
	"go.uber.org/zap"
)

func (us *UpdateService) UpdateNetworkHistorySegments(ctx context.Context, apiURLs []string) error {
	logger := us.log.With(zap.String(UpdaterType, "UpdateNetworkHistorySegments"))

	segmentStore := us.storeService.NewNetworkHistorySegment()

	var (
		successCount, failCount int64
	)

	if len(apiURLs) > 0 && len(us.latestSegmentsCache) < 1 {
		earliestNodeBlock, err := us.readService.GetEarliestBlockHeight(ctx)
		if err != nil {
			return fmt.Errorf("failed to get earliest node block: %w", err)
		}

		us.latestSegmentsCache, err = us.readService.GetLatestFlushedSegmentsHeights(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest flushed segments from db: %w", err)
		}

		// make sure we have latest segments cache for each data-node we want,
		// and latest block is not older the earliest block available locally
		for _, nodeURL := range apiURLs {
			latestSegmentHeight, exist := us.latestSegmentsCache[nodeURL]
			if !exist || (exist && latestSegmentHeight < earliestNodeBlock) {
				us.latestSegmentsCache[nodeURL] = earliestNodeBlock
			}
		}
	}

	latestLocalBlock, err := us.readService.GetLatestLocalBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest local block height: %w", err)
	}

	latestSegmentsToFlushHeight := map[string]int64{}

	latestSegmentInSet := func(segments []*datanode.NetworkHistorySegment) int64 {
		var maxHeight int64 = 0

		for _, segment := range segments {
			if segment.Height > maxHeight {
				maxHeight = segment.Height
			}
		}

		return maxHeight
	}

	for _, apiURL := range apiURLs {
		latestFlushedSegmentHeight, exists := us.latestSegmentsCache[apiURL]
		if !exists {
			us.log.Errorf("cache for latest flushed segment height for %s does not exist, but it should", apiURL)
			continue
		}

		logger.Debug("fetching network history segments", zap.String("url", apiURL))
		dataNodeClient := datanode.NewDataNodeClient(apiURL)
		// We are not interested in the segments outside of the range we are replaying
		segments, err := dataNodeClient.GetNetworkHistorySegments(latestFlushedSegmentHeight, latestLocalBlock)
		if err != nil {
			// Below line is not error of our program. It is one of the expected states in the external data-nodes.
			// Failure of external data-node is valid state and we log in in the data base. We MUST NOT report it as an error
			us.log.Debug("Failed to get Network History segments", zap.String("data-node", apiURL), zap.Error(err))
			failCount += 1
			continue
		}

		for _, segment := range segments {
			segmentStore.AddWithoutTime(segment)
		}

		logger.Debug(fmt.Sprintf("fetched %d network history segments", len(segments)), zap.String("url", apiURL))
		successCount += 1

		latestSegmentsToFlushHeight[apiURL] = latestSegmentInSet(segments)
	}

	storedData, err := segmentStore.FlushUpsertWithoutTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to flush network history segments: %w", err)
	}
	logger.Debug(
		"Stored Segment data in SQLStore",
		zap.Int64("data-node success", successCount),
		zap.Int64("data-node fail", failCount),
		zap.Int("row count", len(storedData)),
	)

	// update local cache only if no error - it means transaction has been committed
	for apiUrl, latestFlushedSegmentHeight := range latestSegmentsToFlushHeight {
		if latestFlushedSegmentHeight > 0 {
			us.latestSegmentsCache[apiUrl] = latestFlushedSegmentHeight
		}
	}

	return nil
}
