package nodescanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dnapipb "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func CheckREST(address string) (time.Duration, uint64, error) {
	var score uint64 = 0
	s, err := url.JoinPath(address, "api/v2/info")
	if err != nil {
		return time.Hour, 0, err
	}
	if strings.HasPrefix(s, "https://") {
		score += 1
	}

	now := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)
	if err != nil {
		return time.Hour, 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Hour, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return time.Hour, 0, fmt.Errorf("unexpected http status code: %v", resp.StatusCode)
	}
	score += 1
	return time.Since(now), score, err
}

func CheckGQL(address string) (time.Duration, uint64, error) {
	var score uint64 = 0
	s := address
	if strings.HasPrefix(s, "https://") {
		score += 1
	}

	now := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s, bytes.NewBuffer([]byte(`{"query": "{epoch{id}}"}`)))
	if err != nil {
		return time.Hour, 0, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Hour, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return time.Hour, 0, fmt.Errorf("unexpected http status code: %v", resp.StatusCode)
	}
	score += 1

	return time.Since(now), score, err
}

func CheckGRPC(address string) (time.Duration, uint64, error) {
	var score uint64 = 0
	useTLS := strings.HasPrefix(address, "tls://")

	var creds credentials.TransportCredentials
	if useTLS {
		address = address[6:]
		creds = credentials.NewClientTLSFromCert(nil, "")
		score += 1
	} else {
		creds = insecure.NewCredentials()
	}

	connection, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return time.Hour, 0, err
	}
	defer connection.Close()

	now := time.Now()

	connDT := dnapipb.NewTradingDataServiceClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = connDT.Info(ctx, &dnapipb.InfoRequest{})
	if err != nil {
		return time.Hour, 0, err
	}
	score += 1

	return time.Since(now), score, err
}

func CheckDataDepth(ctx context.Context, address string) (uint64, uint64, uint64, error) {
	dayAgo := time.Now().Add(-24 * time.Hour)
	weekAgo := time.Now().Add(-6.5 * 24 * time.Hour)
	firstTrade := time.Date(2023, 5, 23, 16, 0, 0, 0, time.UTC)

	count, err := GetTradesCount(ctx, address, dayAgo)
	if err != nil || count == 0 {
		return 0, 0, 0, err
	}
	count, err = GetTradesCount(ctx, address, weekAgo)
	if err != nil || count == 0 {
		return 1, 0, 0, err
	}
	count, err = GetTradesCount(ctx, address, firstTrade)
	if err != nil || count == 0 {
		return 1, 1, 0, err
	}

	return 1, 1, 1, nil
}

func GetTradesCount(ctx context.Context, address string, dateRangeEnd time.Time) (uint64, error) {
	s, err := url.JoinPath(address, "api/v2/trades")
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)
	if err != nil {
		return 0, err
	}
	q := req.URL.Query()
	q.Add("dateRange.endTimestamp", strconv.FormatInt(dateRangeEnd.UTC().UnixNano(), 10))
	q.Add("pagination.first", "1")
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected http status code: %v", resp.StatusCode)
	}
	payload := struct {
		Trades struct {
			Edges []json.RawMessage
		}
	}{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&payload); err != nil {
		return 0, fmt.Errorf("failed to parse trades response: %w", err)
	}
	return uint64(len(payload.Trades.Edges)), nil
}
