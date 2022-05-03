package grpc

import (
	"context"

	"github.com/hm-edu/pki-service/pkg/database"
	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthChecker is the basic structure of a GRPC health check.
type HealthChecker struct{}

// Check performs a single health check.
func (s *HealthChecker) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	zap.L().Debug("Serving the Check request for health check")
	err := database.DB.Internal.PingContext(ctx)
	if err != nil {
		zap.L().Error("Failed to ping the database", zap.Error(err))
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch implements the Watch method of the HealthServer interface.
func (s *HealthChecker) Watch(_ *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	zap.L().Debug("Serving the Watch request for health check")
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

// NewHealthChecker returns a new HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}
