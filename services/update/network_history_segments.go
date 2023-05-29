package update

import (
	"context"

	"github.com/vegaprotocol/data-metrics-store/clients/datanode"
	"go.uber.org/zap"
)

func (us *UpdateService) UpdateNetworkHistorySegments(ctx context.Context, apiURLs []string) error {
	segmentStore := us.storeService.NewNetworkHistorySegment()

	var successCount, failCount int64

	for _, apiURL := range apiURLs {
		dataNodeClient := datanode.NewDataNodeClient(apiURL)

		segments, err := dataNodeClient.GetNetworkHistorySegments()
		if err != nil {
			us.log.Error("Failed to get Network History segments", zap.String("data-node", apiURL), zap.Error(err))
			failCount += 1
			continue
		}

		for _, segment := range segments {
			segmentStore.AddWithoutTime(&segment)
		}
		successCount += 1
	}

	storedData, err := segmentStore.FlushUpsertWithoutTime(ctx)
	if err != nil {
		return err
	}
	us.log.Info(
		"Stored Segment data in SQLStore",
		zap.Int64("data-node success", successCount),
		zap.Int64("data-node fail", failCount),
		zap.Int("row count", len(storedData)),
	)

	return nil
}
