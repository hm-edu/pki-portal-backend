package grpc

import (
	"fmt"
	"net"
	"net/http"
	"time"

	pb "github.com/hm-edu/portal-apis"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/pkg/cfg"
	"github.com/hm-edu/pki-service/pkg/worker"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/hm-edu/sectigo-client/sectigo"

	"github.com/go-co-op/gocron"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger     *zap.Logger
	config     *Config
	sectigoCfg *cfg.SectigoConfiguration
	db         *ent.Client
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-rest-interface-name"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger, sectigoCfg *cfg.SectigoConfiguration, db *ent.Client) (*Server, error) {
	srv := &Server{
		logger:     logger,
		sectigoCfg: sectigoCfg,
		config:     config,
		db:         db,
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
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(),
				tracing.NewGRPUnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(s.logger,
					grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
						if fullMethodName == "/grpc.health.v1.Health/Check" && err == nil {
							return false
						}
						return true
					})))),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(),
				tracing.NewGRPCStreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(s.logger))))

	server := NewHealthChecker()
	reflection.Register(srv)

	// According to https://go.dev/src/net/http/client.go:
	// "Clients are safe for concurrent use by multiple goroutines."
	// => one http client is fine ;)

	c := sectigo.NewClient(http.DefaultClient, s.logger, s.sectigoCfg.User, s.sectigoCfg.Password, s.sectigoCfg.CustomerURI)
	syncer := worker.Syncer{Client: c, Db: s.db}
	cron := gocron.NewScheduler(time.UTC)
	_, err = cron.Every("5m").Do(syncer.SyncPendingCertificates)
	if err != nil {
		s.logger.Warn("failed to schedule syncer", zap.Error(err))
	}
	cron.StartAsync()

	ssl := newSslAPIServer(c, s.sectigoCfg, s.db)
	pb.RegisterSSLServiceServer(srv, ssl)
	smime := newSmimeAPIServer(c, s.sectigoCfg)
	pb.RegisterSmimeServiceServer(srv, smime)
	grpc_health_v1.RegisterHealthServer(srv, server)

	go func() {
		if err := srv.Serve(listener); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}

	}()
	<-stopCh
	s.logger.Info("Stopping GRPC server")
	srv.GracefulStop()
}
