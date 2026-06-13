//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/labstack/gommon/log"

	"github.com/hm-edu/pki-rest-interface/pkg/api/smime"
	"github.com/hm-edu/pki-rest-interface/pkg/api/ssl"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	commonInterceptor "github.com/hm-edu/portal-common/interceptor"
	"github.com/hm-edu/portal-common/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	healthy int32
	ready   int32
)

// @title PKI Service
// @version 2.0
// @description Go microservice for PKI management. Provides an wrapper above the sectigo API.

// @contact.name Source Code
// @contact.url  https://github.com/hm-edu/portal-backend

// @license.name Apache License
// @license.url https://github.com/hm-edu/portal-backend/blob/main/LICENSE

// @securitydefinitions.apikey API
// @in header
// @name Authorization

// Server represents the basic structure of the REST-API server
type Server struct {
	app        *echo.Echo
	logger     *zap.Logger
	config     *api.Config
	handlerCfg *cfg.HandlerConfiguration
}

// NewServer creates a new server
func NewServer(logger *zap.Logger, config *api.Config, handlerCfg *cfg.HandlerConfiguration) *Server {
	return &Server{app: echo.New(), logger: logger, config: config, handlerCfg: handlerCfg}
}

func (server *Server) wireRoutesAndMiddleware() {
	jwks, err := keyfunc.NewDefault([]string{server.config.JwksURI})
	if err != nil {
		server.logger.Fatal("fetching jwk set failed", zap.Error(err))
	}

	config := commonAuth.JWTConfig{
		ParseTokenFunc: func(auth string, _ *echo.Context) (interface{}, error) {
			return commonAuth.GetToken(auth, jwks, server.config.Audience)
		},
	}

	jwtMiddleware := commonAuth.JWTWithConfig(config)

	if server.config.SentryDSN != "" {
		if client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn: server.config.SentryDSN,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate: 1.0,
			EnableTracing:    true,
		}); err != nil {
			log.Warnf("Sentry initialization failed: %v\n", err)
		} else {
			sentry.CurrentHub().BindClient(client)
			cfg := zapsentry.Configuration{
				Level:             zapcore.WarnLevel, //when to send message to sentry
				EnableBreadcrumbs: true,              // enable sending breadcrumbs to Sentry
				BreadcrumbLevel:   zapcore.InfoLevel, // at what level should we sent breadcrumbs to sentry, this level can't be higher than `Level`
			}
			core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))
			if err != nil {
				log.Error("Sentry initialization failed", zap.Error(err))
			}
			server.logger = zapsentry.AttachCoreToLogger(core, server.logger)
			server.app.Use(sentryecho.New(sentryecho.Options{}))
		}
	}
	server.app.Use(middleware.RequestID())
	server.app.Use(logging.ZapLogger(server.logger))
	server.app.Use(middleware.Recover())

	if len(server.config.CorsAllowedOrigins) != 0 {
		server.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     server.config.CorsAllowedOrigins,
			AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "sentry-trace", "baggage"},
			AllowCredentials: false,
			AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodPost, http.MethodDelete},
		}))
	}
	server.app.GET("/healthz", server.healthzHandler)
	server.app.GET("/readyz", server.readyzHandler)
	server.app.GET("/whoami", server.whoamiHandler, jwtMiddleware)

	group := server.app.Group("/ssl")
	{
		domainClient, err := domainClient(server.handlerCfg.DomainService, server.config.SentryDSN)
		if err != nil {
			server.logger.Fatal("failed to create domain client", zap.Error(err))
		}

		sslClient, err := sslClient(server.handlerCfg.SslService, server.config.SentryDSN)
		if err != nil {
			server.logger.Fatal("failed to create ssl client", zap.Error(err))
		}
		ssl := ssl.NewHandler(domainClient, sslClient)
		group.Use(jwtMiddleware)
		group.Use(commonAuth.HasScope("Certificates"))
		group.GET("/", ssl.List)
		group.GET("/active", ssl.Active)
		group.POST("/revoke", ssl.Revoke)
		group.POST("/csr", ssl.HandleCsr)
	}

	group = server.app.Group("/smime")
	{
		smimeClient, err := smimeClient(server.handlerCfg.SmimeService, server.config.SentryDSN)
		if err != nil {
			server.logger.Fatal("failed to create smime client", zap.Error(err))
		}
		handler := smime.NewHandler(smimeClient, server.handlerCfg.RejectStudents)
		group.Use(jwtMiddleware)
		group.Use(commonAuth.HasScope("Certificates"))
		group.GET("/", handler.List)
		group.POST("/revoke", handler.Revoke)
		group.POST("/csr", handler.HandleCsr)
	}
	ready = 1
	healthy = 1
}

func domainClient(host string, sentryDSN string) (pb.DomainServiceClient, error) {
	var interceptor []grpc.UnaryClientInterceptor
	if sentryDSN != "" {
		interceptor = append(interceptor, commonInterceptor.UnaryClientInterceptor())
	}
	conn, err := api.ConnectGRPC(host, grpc.WithChainUnaryInterceptor(interceptor...))
	if err != nil {
		return nil, err
	}
	return pb.NewDomainServiceClient(conn), nil
}

func smimeClient(host string, sentryDSN string) (pb.SmimeServiceClient, error) {
	var interceptor []grpc.UnaryClientInterceptor
	if sentryDSN != "" {
		interceptor = append(interceptor, commonInterceptor.UnaryClientInterceptor())
	}
	conn, err := api.ConnectGRPC(host, grpc.WithChainUnaryInterceptor(interceptor...))
	if err != nil {
		return nil, err
	}
	return pb.NewSmimeServiceClient(conn), nil
}

func sslClient(host string, sentryDSN string) (pb.SSLServiceClient, error) {
	var interceptor []grpc.UnaryClientInterceptor
	if sentryDSN != "" {
		interceptor = append(interceptor, commonInterceptor.UnaryClientInterceptor())
	}
	conn, err := api.ConnectGRPC(host, grpc.WithChainUnaryInterceptor(interceptor...))
	if err != nil {
		return nil, err
	}
	return pb.NewSSLServiceClient(conn), nil
}

// ListenAndServe starts the http server and waits for the channel to stop the server
func (server *Server) ListenAndServe(stopCh <-chan struct{}) {

	server.wireRoutesAndMiddleware()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		addr := server.config.Host + ":" + server.config.Port
		server.logger.Info("Starting HTTP Server.", zap.String("addr", addr))
		sc := echo.StartConfig{
			Address:    addr,
			HideBanner: true,
			HidePort:   true,
		}
		if err := sc.Start(ctx, server.app); err != nil && !errors.Is(err, http.ErrServerClosed) {
			server.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()
	<-stopCh
	cancel()
}
