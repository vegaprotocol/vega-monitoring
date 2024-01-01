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

	var successCount, failCount int64

	for _, apiURL := range apiURLs {
		logger.Debug("fetching network history segments", zap.String("url", apiURL))
		dataNodeClient := datanode.NewDataNodeClient(apiURL)

		segments, err := dataNodeClient.GetNetworkHistorySegments()
		if err != nil {
			us.log.Error("Failed to get Network History segments", zap.String("data-node", apiURL), zap.Error(err))
			failCount += 1
			continue
		}

		for _, segment := range segments {
			segmentStore.AddWithoutTime(segment)
		}

		logger.Debug(fmt.Sprintf("fetched %d network history segments", len(segments)), zap.String("url", apiURL))
		successCount += 1
	}

	storedData, err := segmentStore.FlushUpsertWithoutTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to flush network history segments: %w", err)
	}
	logger.Info(
		"Stored Segment data in SQLStore",
		zap.Int64("data-node success", successCount),
		zap.Int64("data-node fail", failCount),
		zap.Int("row count", len(storedData)),
	)

	return nil
}
