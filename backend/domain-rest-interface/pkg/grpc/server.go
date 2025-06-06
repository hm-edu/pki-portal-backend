package grpc

import (
	"fmt"
	"net"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"

	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server is the basic structure of the GRPC server.
type Server struct {
	logger *zap.Logger
	config *Config
	store  *store.DomainStore
	admins []string
}

// Config is the basic structure of the GRPC configuration
type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-rest-interface-name"`
	SentryDSN   string `mapstructure:"sentry_dsn"`
}

// NewServer creates a new GRPC server
func NewServer(config *Config, logger *zap.Logger, store *store.DomainStore, admins []string) (*Server, error) {
	srv := &Server{
		logger: logger,
		config: config,
		store:  store,
		admins: admins,
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

	interceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(s.logger,
			grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
				if fullMethodName == "/grpc.health.v1.Health/Check" && err == nil {
					return false
				}
				return true
			}),
		)}

	if s.config.SentryDSN != "" {
		if client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn: s.config.SentryDSN,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate:   1.0,
			EnableTracing:      true,
			IgnoreTransactions: []string{"/grpc.health.v1.Health/Check"},
		}); err != nil {
			s.logger.Sugar().Warnf("Sentry initialization failed: %v\n", err)
		} else {
			cfg := zapsentry.Configuration{
				Level:             zapcore.WarnLevel, //when to send message to sentry
				EnableBreadcrumbs: true,              // enable sending breadcrumbs to Sentry
				BreadcrumbLevel:   zapcore.InfoLevel, // at what level should we sent breadcrumbs to sentry, this level can't be higher than `Level`
			}
			core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))
			if err != nil {
				s.logger.Error("Sentry initialization failed", zap.Error(err))
			}
			s.logger = zapsentry.AttachCoreToLogger(core, s.logger)
			interceptors = append(interceptors, grpc_sentry.UnaryServerInterceptor())
		}
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				interceptors...,
			)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(s.logger))))

	server := health.NewServer()
	reflection.Register(srv)
	api := newDomainAPIServer(s.store, s.logger, s.admins)
	pb.RegisterDomainServiceServer(srv, api)
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
