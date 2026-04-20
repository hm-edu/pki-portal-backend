//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/hm-edu/eab-rest-interface/pkg/api/eab"
	commonApi "github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/logging"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ready int32
)

// @title EAB Service
// @version 2.0
// @description Go microservice for EAB management.

// @contact.name Source Code
// @contact.url  https://github.com/hm-edu/portal-backend

// @license.name Apache License
// @license.url https://github.com/hm-edu/portal-backend/blob/main/LICENSE

// @securitydefinitions.apikey API
// @in header
// @name Authorization

// Server is the basic structure of the users' REST-API server
type Server struct {
	app           *echo.Echo
	logger        *zap.Logger
	config        *commonApi.Config
	provisionerID string
}

// NewServer creates a new server
func NewServer(logger *zap.Logger, config *commonApi.Config, provisionerID string) *Server {
	return &Server{app: echo.New(), logger: logger, config: config, provisionerID: provisionerID}
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
	server.app.Use(logging.ZapLogger(server.logger, logging.WithSkipper(func(c *echo.Context) bool {
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
	server.app.GET("/healthz", server.healthzHandler)
	server.app.GET("/readyz", server.readyzHandler)
	server.app.GET("/whoami", server.whoamiHandler, jwtMiddleware)

	v1 := server.app.Group("/eab")
	{
		h := eab.NewHandler(server.provisionerID)
		v1.Use(jwtMiddleware)
		v1.Use(commonAuth.HasScope("EAB"))
		v1.GET("/", h.GetExternalAccountKeys)
		v1.POST("/", h.CreateNewKey)
		v1.DELETE("/:id", h.DeleteKey)
	}

}

// ListenAndServe starts the http server and waits for the channel to stop the server.
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
		if err := sc.Start(ctx, server.app); err != nil && err != http.ErrServerClosed {
			server.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()

	ready = 1

	<-stopCh
	server.logger.Info("Stopping HTTP Server.")
	cancel()
}
