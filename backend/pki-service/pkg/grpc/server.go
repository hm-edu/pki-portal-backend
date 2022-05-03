package grpc

import (
	"fmt"
	"net"

	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/sectigo-client/sectigo"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/pkg/cfg"
	"github.com/hm-edu/portal-common/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger        *zap.Logger
	config        *Config
	sectigoCfg    *cfg.SectigoConfiguration
	sectigoClient *sectigo.Client
	db            *ent.Client
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-rest-interface-name"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger, sectigoCfg *cfg.SectigoConfiguration, sectigoClient *sectigo.Client, db *ent.Client) (*Server, error) {
	srv := &Server{
		logger:        logger,
		sectigoCfg:    sectigoCfg,
		sectigoClient: sectigoClient,
		config:        config,
		db:            db,
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
	var srv *grpc.Server
	srv = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(), tracing.NewGRPUnaryServerInterceptor(), grpc_zap.UnaryServerInterceptor(s.logger))),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(), tracing.NewGRPCStreamServerInterceptor(), grpc_zap.StreamServerInterceptor(s.logger))))

	server := NewHealthChecker()
	reflection.Register(srv)

	ssl := newSslAPIServer(s.sectigoClient, s.sectigoCfg, s.db)
	pb.RegisterSSLServiceServer(srv, ssl)
	smime := newSmimeAPIServer(s.sectigoClient, s.sectigoCfg)
	pb.RegisterSmimeServiceServer(srv, smime)
	grpc_health_v1.RegisterHealthServer(srv, server)

	go func() {
		if err := srv.Serve(listener); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}

	}()
	_ = <-stopCh
	s.logger.Info("Stopping GRPC server")
	srv.GracefulStop()
}
