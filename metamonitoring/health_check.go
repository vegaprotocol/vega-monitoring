package metamonitoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/google/uuid"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/entities"
	"github.com/vegaprotocol/vega-monitoring/services/read"
	"go.uber.org/zap"
)

const responseCacheTime = 30 * time.Second

type readService interface {
	GetMetaMonitoringStatusesExtended(context.Context) (*read.MetaMonitoringStatusesExtended, error)
}

type healthCheckStatusDetails struct {
	Healthy         bool
	UpdatedAt       time.Time
	UnhealthyReason string
}

func newHealthCheckStatusDetailsFromReadStatusDetails(details read.StatusDetails) healthCheckStatusDetails {
	return healthCheckStatusDetails{
		Healthy:         details.Healthy,
		UpdatedAt:       details.UpdatedAt,
		UnhealthyReason: entities.UnHealthyReasonString(details.UnhealthyReason),
	}
}

type healthCheckResponseDetails struct {
	DataNodeData               healthCheckStatusDetails
	AssetPricesData            healthCheckStatusDetails
	BlockSignersData           healthCheckStatusDetails
	CometTxsData               healthCheckStatusDetails
	NetworkBalancesData        healthCheckStatusDetails
	NetworkHistorySegmentsData healthCheckStatusDetails
	GrafanaServer              *healthCheckStatusDetails
}

type healthCheckResponse struct {
	Healthy bool
	Details healthCheckResponseDetails
}

type HealthCheckService struct {
	readService  readService
	mut          sync.Mutex
	cachedAt     time.Time
	config       config.HealthCheckConfig
	lastResponse *healthCheckResponse
	logger       *logging.Logger
}

func NewHealthCheckService(cfg config.HealthCheckConfig, readService readService, logger *logging.Logger) (*HealthCheckService, error) {
	return &HealthCheckService{
		readService: readService,
		config:      cfg,
		logger:      logger.Named("health-check"),
		cachedAt:    time.Unix(0, 0),
	}, nil
}

func (hc *HealthCheckService) fetchGrafanaStatus() *healthCheckStatusDetails {
	resp, err := http.Get(fmt.Sprintf("%s/api/health", strings.TrimRight(hc.config.GrafanaServer.URI, "/")))
	if err != nil {
		hc.logger.Warn("Grafana Server is unhealthy: error during get call", zap.Error(err))
		return &healthCheckStatusDetails{
			Healthy:         false,
			UpdatedAt:       time.Now(),
			UnhealthyReason: entities.UnHealthyReasonString(entities.ReasonTargetConnectionFailure),
		}
	}

	if resp.StatusCode != http.StatusOK {
		hc.logger.Warningf("Grafana Server is unhealthy. Expected %d status code, got %d", http.StatusOK, resp.StatusCode)

		return &healthCheckStatusDetails{
			Healthy:         false,
			UpdatedAt:       time.Now(),
			UnhealthyReason: entities.UnHealthyReasonString(entities.ReasonTargetConnectionFailure),
		}
	}

	return &healthCheckStatusDetails{
		Healthy:         true,
		UpdatedAt:       time.Now(),
		UnhealthyReason: entities.UnHealthyReasonString(entities.ReasonUnknown),
	}
}

func (hc *HealthCheckService) fetchStatus(ctx context.Context) (*healthCheckResponse, error) {
	statuses, err := hc.readService.GetMetaMonitoringStatusesExtended(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest monitoring statuses: %w", err)
	}

	var grafanaServerStatus *healthCheckStatusDetails
	if hc.config.GrafanaServer.Enabled {
		grafanaServerStatus = hc.fetchGrafanaStatus()
	}

	return &healthCheckResponse{
		Healthy: statuses.HealthyOverAll && (grafanaServerStatus == nil || grafanaServerStatus.Healthy),
		Details: healthCheckResponseDetails{
			DataNodeData:               newHealthCheckStatusDetailsFromReadStatusDetails(statuses.DataNodeData),
			AssetPricesData:            newHealthCheckStatusDetailsFromReadStatusDetails(statuses.AssetPricesData),
			BlockSignersData:           newHealthCheckStatusDetailsFromReadStatusDetails(statuses.BlockSignersData),
			CometTxsData:               newHealthCheckStatusDetailsFromReadStatusDetails(statuses.CometTxsData),
			NetworkBalancesData:        newHealthCheckStatusDetailsFromReadStatusDetails(statuses.NetworkBalancesData),
			NetworkHistorySegmentsData: newHealthCheckStatusDetailsFromReadStatusDetails(statuses.NetworkHistorySegmentsData),
			GrafanaServer:              grafanaServerStatus,
		},
	}, nil
}

func (hc *HealthCheckService) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hc.mut.Lock()
		defer hc.mut.Unlock()
		requestId := uuid.New().String()

		now := time.Now()
		var err error
		if hc.lastResponse == nil || now.Sub(hc.cachedAt) >= responseCacheTime {
			hc.lastResponse, err = hc.fetchStatus(context.Background())
			hc.cachedAt = time.Now()
			if err != nil {
				hc.logger.Error("Failed to fetch monitoring status", zap.Error(err), zap.String("requestId", requestId))
				if _, err := w.Write([]byte(fmt.Sprintf("Internal server error (request %s)", requestId))); err != nil {
					hc.logger.Error("failed to write 5XX for fetch Status", zap.Error(err))
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		response, err := json.MarshalIndent(hc.lastResponse, "", "    ")
		if err != nil {
			hc.logger.Error("Failed to marshal response", zap.Error(err), zap.String("requestId", requestId))
			if _, err := w.Write([]byte(fmt.Sprintf("Internal server error (request %s)", requestId))); err != nil {
				hc.logger.Error("failed to write 5XX response for unmarshal: %w", zap.Error(err))
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(response); err != nil {
			hc.logger.Error("failed to write healthy/unhealthy response: %w", zap.Error(err))
		}

		if hc.lastResponse.Healthy {
			// Implicitly set status OK
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (hc *HealthCheckService) Run(ctx context.Context, port int) error {
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        hc.handler(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func(server *http.Server) {
		hc.logger.Infof("Starting server at port %d", port)
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
		}
	}(srv)

	time.Sleep(3 * time.Second)
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
