package nodescanner

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vegaprotocol/vega-monitoring/clients/datanode"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

func requestDataNodeStats(client *datanode.DataNodeClient) (*types.DataNodeStatus, error) {
	// Request Core statistics - on data-node endpoint they should contain data-node headers
	coreStatus, headers, err := requestCoreStats(client, []string{"x-block-height", "x-block-timestamp"})
	if err != nil {
		return nil, err
	}

	// parse data-node headers
	strDataNodeBlockHeight := headers["x-block-height"]
	if len(strDataNodeBlockHeight) == 0 {
		return nil, fmt.Errorf("failed to get x-block-height response header: empty header value")
	}
	dataNodeBlockHeight, err := strconv.ParseUint(strDataNodeBlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x-block-height response header %s, %w", strDataNodeBlockHeight, err)
	}

	strDataNodeTime := headers["x-block-timestamp"]
	if len(strDataNodeTime) == 0 {
		return nil, fmt.Errorf("failed to get x-block-timestamp response header, %w", err)
	}
	intDataNodeTime, err := strconv.ParseInt(strDataNodeTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x-block-timestamp response header to int %s, %w", strDataNodeTime, err)
	}
	dataNodeTime := time.Unix(intDataNodeTime/int64(time.Nanosecond), 0)

	return &types.DataNodeStatus{
		CoreStatus:          *coreStatus,
		DataNodeTime:        dataNodeTime,
		DataNodeBlockHeight: dataNodeBlockHeight,
	}, nil
}

func getUnhealthyDataNodeStats() *types.DataNodeStatus {
	return &types.DataNodeStatus{
		CoreStatus:          *getUnhealthyCoreStats(),
		DataNodeTime:        time.Unix(0, 0),
		DataNodeBlockHeight: 0,
	}
}
