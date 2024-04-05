package datanode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/vegaprotocol/vega-monitoring/entities"
)

const DataNodeHeightHeader = "x-block-height"

var (
	ErrHttpCallError            = errors.New("error during http call")
	ErrMissingOrInvalidResponse = errors.New("missing some information from the data-node response")
	ErrTimeGapTooBig            = errors.New("time gap between core and data-node is too big")
	ErrBlocksGapTooBig          = errors.New("blocks gap between core and data-node is too big")
)

func (c *DataNodeClient) IsHealthy(ctx context.Context) (bool, error) {
	payload, headers, err := c.GetStatisticsWithHeaders(ctx, []string{DataNodeHeightHeader})
	if err != nil {
		return false, errors.Join(ErrHttpCallError, err)
	}

	if len(headers[DataNodeHeightHeader]) < 1 {
		return false, errors.Join(ErrMissingOrInvalidResponse, fmt.Errorf("missing %s header in the response", DataNodeHeightHeader))
	}

	currentTime, err := time.Parse(time.RFC3339, payload.CurrentTime)
	if err != nil {
		return false, errors.Join(ErrMissingOrInvalidResponse, fmt.Errorf("failed to parse currentTime %s: %w", payload.CurrentTime, err))
	}
	vegaTime, err := time.Parse(time.RFC3339, payload.VegaTime)
	if err != nil {
		return false, errors.Join(ErrMissingOrInvalidResponse, fmt.Errorf("failed to parse vegaTime %s: %w", payload.VegaTime, err))
	}

	strDataNodeBlockHeight := headers["x-block-height"]
	dataNodeBlockHeight, err := strconv.ParseUint(strDataNodeBlockHeight, 10, 64)
	if err != nil {
		return false, errors.Join(ErrMissingOrInvalidResponse, fmt.Errorf("failed to parse x-block-height response header %s: %w", strDataNodeBlockHeight, err))
	}

	if math.Abs(float64(payload.BlockHeight-dataNodeBlockHeight)) > 150 {
		return false, errors.Join(ErrBlocksGapTooBig, fmt.Errorf("gap between data node and core is bigger than 150 blocks"))
	}

	if currentTime.Sub(vegaTime) > time.Second*120 {
		return false, errors.Join(ErrTimeGapTooBig, fmt.Errorf("gap between data node and core is bigger than 120 seconds"))
	}

	return true, nil
}

func (c *DataNodeClient) GetStatistics(ctx context.Context) (*entities.Statistics, error) {
	resp, _, err := c.GetStatisticsWithHeaders(ctx, nil)

	return resp, err
}

func (c *DataNodeClient) GetStatisticsWithHeaders(ctx context.Context, wantedHeaders []string) (*entities.Statistics, map[string]string, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, nil, errors.Join(errWaitingForRateLimiter, fmt.Errorf("failed to get network history segments for %s: %w", c.apiURL, err))
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(statisticsURL, c.apiURL),
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new request with context: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to do http request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("invalid response code for statistics request: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var payload struct {
		Statistics entities.Statistics `json:"statistics"`
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal statistics response: %w", err)
	}

	headerValues := map[string]string{}

	for _, header := range wantedHeaders {
		headerValues[header] = resp.Header.Get(header)
	}

	return &payload.Statistics, headerValues, nil
}
