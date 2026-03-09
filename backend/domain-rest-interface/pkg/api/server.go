//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/hm-edu/domain-rest-interface/pkg/api/docs"
	"github.com/hm-edu/domain-rest-interface/pkg/api/domains"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	commonApi "github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/logging"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	echoSwagger "github.com/swaggo/echo-swagger"

	// Required for the generation of swagger docs
	_ "github.com/hm-edu/domain-rest-interface/pkg/api/docs"
)

var (
	ready       int32
	openAPISpec *openapi3.T
)

// @title Domain Service
// @version 2.0
// @description Go microservice for Domain management.

// @contact.name Source Code
// @contact.url  https://github.com/hm-edu/portal-backend

// @license.name Apache License
// @license.url https://github.com/hm-edu/portal-backend/blob/main/LICENSE

// @securitydefinitions.apikey API
// @in header
// @name Authorization

// Server is the basic structure of the users' REST-API server
type Server struct {
	app        *echo.Echo
	logger     *zap.Logger
	config     *commonApi.Config
	store      *store.DomainStore
	pkiSerivce pb.SSLServiceClient
	admins     []string
}

// NewServer creates a new server
func NewServer(logger *zap.Logger, config *commonApi.Config, store *store.DomainStore, pkiSerivce pb.SSLServiceClient, admins []string) *Server {

	return &Server{app: echo.New(), logger: logger, config: config, store: store, pkiSerivce: pkiSerivce, admins: admins}
}

func (server *Server) wireRoutesAndMiddleware() {

	server.app.HideBanner = true
	server.app.HidePort = true

	jwks, err := keyfunc.NewDefault([]string{server.config.JwksURI})
	if err != nil {
		server.logger.Fatal("fetching jwk set failed", zap.Error(err))
	}

	config := commonAuth.JWTConfig{
		ParseTokenFunc: func(auth string, _ echo.Context) (interface{}, error) {
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
	server.app.Use(logging.ZapLogger(server.logger, logging.WithSkipper(func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/docs") || strings.Contains(c.Path(), "/healthz")
	})))
	server.app.Use(middleware.Recover())

	if len(server.config.CorsAllowedOrigins) != 0 {
		server.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     server.config.CorsAllowedOrigins,
			AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "sentry-trace", "baggage"},
			AllowCredentials: false,
			AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodPost, http.MethodDelete},
		}))
	}
	server.app.GET("/docs/spec.json", func(c echo.Context) error {
		if openAPISpec == nil {
			spec, err := commonApi.ToOpenAPI3(docs.SwaggerInfo)
			if err != nil {
				return err
			}
			openAPISpec = spec
		}
		return c.JSON(http.StatusOK, openAPISpec)
	})

	server.app.GET("/docs/*", echoSwagger.EchoWrapHandler(func(c *echoSwagger.Config) {
		c.URL = "/docs/spec.json"
	}))
	server.app.GET("/healthz", server.healthzHandler)
	server.app.GET("/readyz", server.readyzHandler)
	server.app.GET("/whoami", server.whoamiHandler, jwtMiddleware)

	v1 := server.app.Group("/domains")
	{
		h := domains.NewHandler(server.store, server.pkiSerivce, server.admins)
		v1.Use(jwtMiddleware)
		v1.Use(commonAuth.HasScope("Domains"))
		v1.GET("/", h.ListDomains)
		v1.POST("/", h.CreateDomain)
		{
			v1.DELETE("/:id", h.DeleteDomain)
			v1.POST("/:id/approve", h.ApproveDomain)
			v1.POST("/:id/transfer", h.TransferDomain)
			v1.POST("/:id/delegation", h.AddDelegation)
			v1.DELETE("/:id/delegation/:delegation", h.DeleteDelegation)
		}
	}

}

// ListenAndServe starts the http server and waits for the channel to stop the server.
func (server *Server) ListenAndServe(stopCh <-chan struct{}) {

	server.wireRoutesAndMiddleware()
	go func() {
		addr := server.config.Host + ":" + server.config.Port
		server.logger.Info("Starting HTTP Server.", zap.String("addr", addr))
		if err := server.app.Start(addr); err != http.ErrServerClosed {
			server.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()

	ready = 1

	<-stopCh
	err := server.app.Shutdown(context.Background())
	if err != nil {
		server.logger.Fatal("Stopping http server failed", zap.Error(err))
	}
	defer sentry.Flush(2 * time.Second)

}
