package nodescanner

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	dnapipb "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func checkREST(address string) (time.Duration, uint64, error) {
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

func checkGQL(address string) (time.Duration, uint64, error) {
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

func checkGRPC(address string) (time.Duration, uint64, error) {
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
