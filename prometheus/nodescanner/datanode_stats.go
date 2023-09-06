package nodescanner

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

func requestDataNodeStats(address string) (*types.DataNodeStatus, error) {

	// Request Core statistics - on data-node endpoint they should contain data-node headers
	coreStatus, headers, err := requestCoreStats(address, []string{"x-block-height", "x-block-timestamp"})
	if err != nil {
		return nil, err
	}

	// parse data-node headers
	strDataNodeBlockHeight := headers["x-block-height"]
	if len(strDataNodeBlockHeight) == 0 {
		return nil, fmt.Errorf("failed to check REST %s, failed to get x-block-height response header, %w", address, err)
	}
	dataNodeBlockHeight, err := strconv.ParseUint(strDataNodeBlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse x-block-height response header %s, %w", address, strDataNodeBlockHeight, err)
	}

	strDataNodeTime := headers["x-block-timestamp"]
	if len(strDataNodeTime) == 0 {
		return nil, fmt.Errorf("failed to check REST %s, failed to get x-block-timestamp response header, %w", address, err)
	}
	intDataNodeTime, err := strconv.ParseInt(strDataNodeTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to check REST %s, failed to parse x-block-timestamp response header to int %s, %w", address, strDataNodeTime, err)
	}
	dataNodeTime := time.Unix(intDataNodeTime, 0)

	return &types.DataNodeStatus{
		CoreStatus:          *coreStatus,
		DataNodeTime:        dataNodeTime,
		DataNodeBlockHeight: dataNodeBlockHeight,
		DataNodeScore:       1,
	}, nil
}

func getUnhealthyDataNodeStats() *types.DataNodeStatus {
	return &types.DataNodeStatus{
		CoreStatus:          *getUnhealthyCoreStats(),
		DataNodeTime:        time.Unix(0, 0),
		DataNodeBlockHeight: 0,
		DataNodeScore:       0,
	}
}
