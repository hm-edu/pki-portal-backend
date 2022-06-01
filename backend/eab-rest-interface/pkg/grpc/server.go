package grpc

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/tracing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger        *zap.Logger
	config        *Config
	provisionerID string
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port          int    `mapstructure:"grpc-port"`
	ServiceName   string `mapstructure:"grpc-rest-interface-name"`
	DomainService string `mapstructure:"domain_service"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger, provisionerID string) (*Server, error) {
	srv := &Server{
		logger:        logger,
		config:        config,
		provisionerID: provisionerID,
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

	srv := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(), tracing.NewGRPUnaryServerInterceptor(), grpc_zap.UnaryServerInterceptor(s.logger))),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(), tracing.NewGRPCStreamServerInterceptor(), grpc_zap.StreamServerInterceptor(s.logger))))

	server := health.NewServer()
	reflection.Register(srv)

	domainClient, err := domainClient(s.config.DomainService)
	if err != nil {
		s.logger.Fatal("failed to create domain client", zap.Error(err))
	}

	api := newEabAPIServer(domainClient, s.logger, s.provisionerID)
	pb.RegisterEABServiceServer(srv, api)
	grpc_health_v1.RegisterHealthServer(srv, server)
	server.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(listener); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}
	}()
	<-stopCh
	s.logger.Info("Stopping GRPC server")
	srv.GracefulStop()
}

func domainClient(host string) (pb.DomainServiceClient, error) {
	conn, err := api.ConnectGRPC(host)
	if err != nil {
		return nil, err
	}
	return pb.NewDomainServiceClient(conn), nil
}
