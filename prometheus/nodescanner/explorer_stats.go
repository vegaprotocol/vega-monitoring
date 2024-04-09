package nodescanner

import (
	"context"
	"fmt"

	"github.com/vegaprotocol/vega-monitoring/clients/blockexplorer"
	"github.com/vegaprotocol/vega-monitoring/clients/datanode"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

func requestBlockExplorerStats(coreClient *datanode.DataNodeClient, beClient *blockexplorer.Client) (*types.BlockExplorerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get core stats
	coreStatus, _, err := requestCoreStats(coreClient, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to request core stats: %w", err)
	}
	payload, err := beClient.GetInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get info about block explorer: %w", err)
	}

	return &types.BlockExplorerStatus{
		CoreStatus:               *coreStatus,
		BlockExplorerVersion:     payload.Version,
		BlockExplorerVersionHash: payload.CommitHash,
	}, nil
}

func getUnhealthyBlockExplorerStats() *types.BlockExplorerStatus {
	return &types.BlockExplorerStatus{
		CoreStatus:               *getUnhealthyCoreStats(),
		BlockExplorerVersion:     "",
		BlockExplorerVersionHash: "",
	}
}
