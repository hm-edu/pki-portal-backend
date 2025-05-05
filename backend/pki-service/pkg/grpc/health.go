package grpc

import (
	"context"

	"github.com/hm-edu/pki-service/pkg/database"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// HealthChecker is the basic structure of a GRPC health check.
type HealthChecker struct{}

// Check performs a single health check.
func (s *HealthChecker) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	zap.L().Debug("Serving the Check request for health check")
	err := database.DB.Internal.PingContext(ctx)
	if err != nil {
		zap.L().Error("Failed to ping the database", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavailable")
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

// List implements the List method of the HealthServer interface.
func (s *HealthChecker) List(_ context.Context, _ *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	return &grpc_health_v1.HealthListResponse{
		Statuses: map[string]*grpc_health_v1.HealthCheckResponse{},
	}, nil
}

// NewHealthChecker returns a new HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}
