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

func (c *DataNodeClient) GetNetworkHistorySegments(fromBlock, toBlock int64) ([]*NetworkHistorySegment, error) {
	response, err := c.requestNetworkHistorySegmets()
	if err != nil {
		return nil, err
	}
	result := []*NetworkHistorySegment{}
	for _, segment := range response.Segments {
		height, err := strconv.ParseInt(segment.ToHeight, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ToHeight %s, %w", segment.ToHeight, err)
		}

		// filter blocks We do not want
		if height < fromBlock || height > toBlock {
			continue
		}

		result = append(result, &NetworkHistorySegment{
			Height:    height,
			SegmentId: segment.SegmentId,
			DataNode:  c.apiURL,
		})
	}

	return result, nil
}
