package grpc

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	creds "google.golang.org/grpc/credentials/xds"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger *zap.Logger
	config *Config
	store  *store.DomainStore
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-rest-interface-name"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger, store *store.DomainStore) (*Server, error) {
	srv := &Server{
		logger: logger,
		config: config,
		store:  store,
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
	creds, err := creds.NewServerCredentials(creds.ServerOptions{FallbackCreds: insecure.NewCredentials()})
	if err != nil {
		s.logger.Fatal("failed to get credentials")
	}
	srv := xds.NewGRPCServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(tracing.NewGRPUnaryServerInterceptor(), grpc_zap.UnaryServerInterceptor(s.logger))),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(tracing.NewGRPCStreamServerInterceptor(), grpc_zap.StreamServerInterceptor(s.logger))), grpc.Creds(creds))

	server := health.NewServer()
	reflection.Register(srv)
	api := newDomainAPIServer(s.store)
	pb.RegisterDomainServiceServer(srv, api)
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