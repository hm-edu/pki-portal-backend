package grpc

import (
	"fmt"
	"net"

	"github.com/getsentry/sentry-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	SentryDSN   string `mapstructure:"sentry_dsn"`
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
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: s.config.SentryDSN,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate:   1.0,
			EnableTracing:      true,
			IgnoreTransactions: []string{"/grpc.health.v1.Health/Check"},
		}); err != nil {
			zap.L().Sugar().Warnf("Sentry initialization failed: %v\n", err)
		} else {
			s.logger.Info("Sentry initialized")
			interceptors = append(interceptors, grpc_sentry.UnaryServerInterceptor())
		}
	}

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
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
