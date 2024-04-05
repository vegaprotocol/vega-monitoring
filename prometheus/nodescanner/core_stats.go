package nodescanner

import (
	"context"
	"fmt"
	"time"

	"github.com/vegaprotocol/vega-monitoring/clients/datanode"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
)

var (
	timeout = 5 * time.Second
)

func requestCoreStats(client *datanode.DataNodeClient, headers []string) (*types.CoreStatus, map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	payload, headerValues, err := client.GetStatisticsWithHeaders(ctx, headers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get statistics for core: %w", err)
	}
	currentTime, err := time.Parse(time.RFC3339, payload.CurrentTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse currentTime %s, %w", payload.CurrentTime, err)
	}
	vegaTime, err := time.Parse(time.RFC3339, payload.VegaTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse vegaTime %s, %w", payload.VegaTime, err)
	}

	return &types.CoreStatus{
			CurrentTime:        currentTime,
			CoreTime:           vegaTime,
			CoreBlockHeight:    payload.BlockHeight,
			CoreChainId:        payload.ChainId,
			CoreAppVersion:     payload.AppVersion,
			CoreAppVersionHash: payload.AppVersionHash,
		},
		headerValues,
		nil
}

func getUnhealthyCoreStats() *types.CoreStatus {
	return &types.CoreStatus{
		CurrentTime:        time.Now().UTC(),
		CoreTime:           time.Unix(0, 0),
		CoreBlockHeight:    0,
		CoreChainId:        "",
		CoreAppVersion:     "",
		CoreAppVersionHash: "",
	}
}
