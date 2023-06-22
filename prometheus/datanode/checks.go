package datanode

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

func checkREST(address string) (time.Duration, error) {
	s, err := url.JoinPath(address, "api/v2/info")
	if err != nil {
		return 0, err
	}

	now := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)
	if err == nil {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return time.Since(now), err
		}
		if resp.StatusCode != http.StatusOK {
			return time.Since(now), fmt.Errorf("unexpected http status code: %v", resp.StatusCode)
		}
	}
	return time.Since(now), err
}

func checkGQL(address string) (time.Duration, error) {
	s := address

	now := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s, bytes.NewBuffer([]byte(`{"query": "{epoch{id}}"}`)))
	if err == nil {
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return time.Since(now), err
		}
		if resp.StatusCode != http.StatusOK {
			return time.Since(now), fmt.Errorf("unexpected http status code: %v", resp.StatusCode)
		}
	}

	return time.Since(now), err
}

func checkGRPC(address string) (time.Duration, error) {
	useTLS := strings.HasPrefix(address, "tls://")

	var creds credentials.TransportCredentials
	if useTLS {
		address = address[6:]
		creds = credentials.NewClientTLSFromCert(nil, "")
	} else {
		creds = insecure.NewCredentials()
	}

	connection, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return 0, err
	}

	now := time.Now()

	connDT := dnapipb.NewTradingDataServiceClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = connDT.Info(ctx, &dnapipb.InfoRequest{})

	return time.Since(now), err
}
