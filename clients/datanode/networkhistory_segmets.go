package datanode

import (
	"fmt"
	"strconv"
)

type NetworkHistorySegment struct {
	Height    int64  `db:"height"`
	SegmentId string `db:"segment_id"`
	DataNode  string `db:"data_node"`
}

func (c *DataNodeClient) GetNetworkHistorySegments() ([]*NetworkHistorySegment, error) {
	response, err := c.requestNetworkHistorySegmets()
	if err != nil {
		return nil, err
	}
	result := make([]*NetworkHistorySegment, len(response.Segments))
	for i, segment := range response.Segments {
		height, err := strconv.ParseInt(segment.ToHeight, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ToHeight %s, %w", segment.ToHeight, err)
		}
		result[i] = &NetworkHistorySegment{
			Height:    height,
			SegmentId: segment.SegmentId,
			DataNode:  c.apiURL,
		}
	}
	return result, nil
}
