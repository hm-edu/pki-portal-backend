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
	"github.com/hm-edu/portal-common/auth"
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
	api.app.Use(logging.ZapLogger(api.logger, logging.WithSkipper(func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/docs") || strings.Contains(c.Path(), "/healthz")
	})))
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
	}))
	api.app.GET("/healthz", api.healthzHandler)
	api.app.GET("/readyz", api.readyzHandler)
	api.app.GET("/whoami", api.whoamiHandler, jwtMiddleware)

	v1 := api.app.Group("/domains")
	{
		h := domains.NewHandler(api.store, api.pkiSerivce, api.admins)
		v1.Use(jwtMiddleware)
		v1.Use(auth.HasScope("Domains"))
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
func (api *Server) ListenAndServe(stopCh <-chan struct{}) {

	api.wireRoutesAndMiddleware()
	go func() {
		addr := api.config.Host + ":" + api.config.Port
		api.logger.Info("Starting HTTP Server.", zap.String("addr", addr))
		if err := api.app.Start(addr); err != http.ErrServerClosed {
			api.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()

	ready = 1

	<-stopCh
	err := api.app.Shutdown(context.Background())
	if err != nil {
		api.logger.Fatal("Stopping http server failed", zap.Error(err))
	}
	defer sentry.Flush(2 * time.Second)

}
