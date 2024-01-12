package read

import (
	"context"
	"fmt"
)

func (s *ReadService) GetLatestFlushedSegmentsHeights(ctx context.Context) (map[string]int64, error) {
	networkHistoryStore := s.storeReadService.NewNetworkHistorySegment()

	latestSegments, err := networkHistoryStore.GetLatestSegmentsPerDataNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest segments from the database: %w")
	}

	result := map[string]int64{}
	for _, segment := range latestSegments {
		result[segment.DataNode] = segment.Height
	}

	return result, nil
}
