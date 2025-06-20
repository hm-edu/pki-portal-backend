//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/hm-edu/pki-rest-interface/pkg/api/docs"
	"github.com/hm-edu/pki-rest-interface/pkg/api/smime"
	"github.com/hm-edu/pki-rest-interface/pkg/api/ssl"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/api"
	commonApi "github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/auth"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/logging"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// Required for the generation of swagger docs
	_ "github.com/hm-edu/pki-rest-interface/pkg/api/docs"
)

var (
	healthy     int32
	ready       int32
	openAPISpec *openapi3.T
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
	config     *commonApi.Config
	handlerCfg *cfg.HandlerConfiguration
}

// NewServer creates a new server
func NewServer(logger *zap.Logger, config *commonApi.Config, handlerCfg *cfg.HandlerConfiguration) *Server {
	return &Server{app: echo.New(), logger: logger, config: config, handlerCfg: handlerCfg}
}

func (api *Server) wireRoutesAndMiddleware() {
	api.app.HideBanner = true
	api.app.HidePort = true

	jwks, err := keyfunc.NewDefault([]string{api.config.JwksURI})
	if err != nil {
		api.logger.Fatal("fetching jwk set failed", zap.Error(err))
	}

	config := auth.JWTConfig{
		ParseTokenFunc: func(auth string, _ echo.Context) (interface{}, error) {
			return commonAuth.GetToken(auth, jwks, api.config.Audience)
		},
	}

	jwtMiddleware := auth.JWTWithConfig(config)

	if api.config.SentryDSN != "" {
		if client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn: api.config.SentryDSN,
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
			api.logger = zapsentry.AttachCoreToLogger(core, api.logger)
			api.app.Use(sentryecho.New(sentryecho.Options{}))
		}
	}
	api.app.Use(middleware.RequestID())
	api.app.Use(logging.ZapLogger(api.logger))
	api.app.Use(middleware.Recover())

	if len(api.config.CorsAllowedOrigins) != 0 {
		api.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     api.config.CorsAllowedOrigins,
			AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "sentry-trace", "baggage"},
			AllowCredentials: false,
			AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodPost, http.MethodDelete},
		}))
	}
	api.app.GET("/docs/spec.json", func(c echo.Context) error {
		if openAPISpec == nil {
			spec, err := commonApi.ToOpenAPI3(docs.SwaggerInfo)
			if err != nil {
				return err
			}
			openAPISpec = spec
		}
		return c.JSON(http.StatusOK, openAPISpec)
	})
	api.app.GET("/docs/*", echoSwagger.EchoWrapHandler(func(c *echoSwagger.Config) {
		c.URLs = []string{"/docs/spec.json"}
	})) // default
	api.app.GET("/healthz", api.healthzHandler)
	api.app.GET("/readyz", api.readyzHandler)
	api.app.GET("/whoami", api.whoamiHandler, jwtMiddleware)

	group := api.app.Group("/ssl")
	{
		domainClient, err := domainClient(api.handlerCfg.DomainService, api.config.SentryDSN)
		if err != nil {
			api.logger.Fatal("failed to create domain client", zap.Error(err))
		}

		sslClient, err := sslClient(api.handlerCfg.SslService, api.config.SentryDSN)
		if err != nil {
			api.logger.Fatal("failed to create ssl client", zap.Error(err))
		}
		ssl := ssl.NewHandler(domainClient, sslClient)
		group.Use(jwtMiddleware)
		group.Use(auth.HasScope("Certificates"))
		group.GET("/", ssl.List)
		group.GET("/active", ssl.Active)
		group.POST("/revoke", ssl.Revoke)
		group.POST("/csr", ssl.HandleCsr)
	}

	group = api.app.Group("/smime")
	{
		smimeClient, err := smimeClient(api.handlerCfg.SmimeService, api.config.SentryDSN)
		if err != nil {
			api.logger.Fatal("failed to create smime client", zap.Error(err))
		}
		handler := smime.NewHandler(smimeClient, api.handlerCfg.RejectStudents)
		group.Use(jwtMiddleware)
		group.Use(auth.HasScope("Certificates"))
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
		interceptor = append(interceptor, grpc_sentry.UnaryClientInterceptor())
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
		interceptor = append(interceptor, grpc_sentry.UnaryClientInterceptor())
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
		interceptor = append(interceptor, grpc_sentry.UnaryClientInterceptor())
	}
	conn, err := api.ConnectGRPC(host, grpc.WithChainUnaryInterceptor(interceptor...))
	if err != nil {
		return nil, err
	}
	return pb.NewSSLServiceClient(conn), nil
}

// ListenAndServe starts the http server and waits for the channel to stop the server
func (api *Server) ListenAndServe(stopCh <-chan struct{}) {

	api.wireRoutesAndMiddleware()
	go func() {
		addr := api.config.Host + ":" + api.config.Port
		api.logger.Info("Starting HTTP Server.", zap.String("addr", addr))
		if err := api.app.Start(addr); !errors.Is(err, http.ErrServerClosed) {
			api.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()
	<-stopCh
	err := api.app.Shutdown(context.Background())
	if err != nil {
		api.logger.Fatal("Stopping http server failed", zap.Error(err))
	}
}
