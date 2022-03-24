package grpc

import (
	"fmt"
	"net"

	pb "github.com/hm-edu/domain-api"
	"github.com/hm-edu/portal-common/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger *zap.Logger
	config *Config
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-service-name"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger) (*Server, error) {
	srv := &Server{
		logger: logger,
		config: config,
	}

	return srv, nil
}

// ListenAndServe starts the GRPC server and waits for requests
func (s *Server) ListenAndServe(stopCh <-chan struct{}) {
	addr := fmt.Sprintf(":%v", s.config.Port)
	s.logger.Info("Starting GRPC Server.", zap.String("addr", addr))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Int("port", s.config.Port))
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(tracing.NewGRPUnaryServerInterceptor()),
		grpc.StreamInterceptor(tracing.NewGRPCStreamServerInterceptor()))

	server := health.NewServer()
	reflection.Register(srv)
	api := domainAPIServer{}
	pb.RegisterDomainServiceServer(srv, &api)
	grpc_health_v1.RegisterHealthServer(srv, server)
	server.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(listener); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}
	}()
	_ = <-stopCh
	s.logger.Info("Stopping GRPC server")
	srv.GracefulStop()
}
