package prometheus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vegaprotocol/vega-monitoring/config"
	vega_collectors "github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
)

type PrometheusService struct {
	config                  *config.PrometheusConfig
	server                  *http.Server
	promHandler             *http.Handler
	VegaMonitoringCollector *vega_collectors.VegaMonitoringCollector
}

func NewPrometheusService(cfg *config.PrometheusConfig) *PrometheusService {
	vegaMonitoringCollector := vega_collectors.NewVegaMonitoringCollector()
	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	prometheus.WrapRegistererWithPrefix("vega_monitoring_", promRegistry).MustRegister(
		vegaMonitoringCollector,
	)
	promHandler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry})

	return &PrometheusService{
		config:                  cfg,
		promHandler:             &promHandler,
		VegaMonitoringCollector: vegaMonitoringCollector,
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
	return fmt.Errorf("failed to run Prometheus Monitoring Http service")
	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to run prometheus http server: %w", err)
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
