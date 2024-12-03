package tracing

import (
	"errors"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// InitTracer performs the initialization of the prometheus endpoint.
func InitTracer(logger *zap.Logger) *http.Server {

	server := &http.Server{ // nolint:gosec
		Addr: ":2222",
	}

	http.Handle("/", promhttp.Handler())
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	logger.Info("Prometheus server running on :2222")
	return server
}
