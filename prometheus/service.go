package prometheus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vegaprotocol/vega-monitoring/config"
)

type PrometheusService struct {
	config      *config.PrometheusConfig
	server      *http.Server
	promHandler *http.Handler
	Metrics     *Metrics
}

func NewPrometheusService(cfg *config.PrometheusConfig) *PrometheusService {
	promRegistry := prometheus.NewRegistry()
	metrics := NewMetrics(promRegistry)
	promHandler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry})

	return &PrometheusService{
		config:      cfg,
		promHandler: &promHandler,
		Metrics:     metrics,
	}
}

func (s *PrometheusService) Start() error {
	// Setup Http Service
	mux := http.NewServeMux()
	mux.Handle(s.config.Path, *s.promHandler)

	// Start Http Service
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}
	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to run Prometheus Monitoring Http service, %w", err)
	}
	return nil
}

func (s *PrometheusService) Shutdown(ctx context.Context) error {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop Prometheus Monitoring Http service, %w", err)
		}
	}
	return nil
}
