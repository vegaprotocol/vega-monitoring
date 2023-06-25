package nodescanner

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vegaprotocol/vega-monitoring/prometheus"
)

var (
	timeout = 2 * time.Second
)

func requestDataNodeStats(address string) (*prometheus.DataNodeChecksResults, error) {
	coreCheckResults, headers, err := requestCoreStats(address, []string{"x-block-height", "x-block-timestamp"})
	if err != nil {
		return nil, err
	}

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

	return &prometheus.DataNodeChecksResults{
		CoreCheckResults:    *coreCheckResults,
		DataNodeTime:        dataNodeTime,
		DataNodeBlockHeight: dataNodeBlockHeight,
	}, nil
}
