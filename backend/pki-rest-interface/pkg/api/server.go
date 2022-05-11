//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lestrrat-go/jwx/jwk"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/hm-edu/pki-rest-interface/pkg/api/docs"
	"github.com/hm-edu/pki-rest-interface/pkg/api/smime"
	"github.com/hm-edu/pki-rest-interface/pkg/api/ssl"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/api"
	commonApi "github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/logging"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

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
	ar := jwk.NewAutoRefresh(context.Background())

	api.app.HideBanner = true
	api.app.HidePort = true

	ar.Configure(api.config.JwksURI, jwk.WithMinRefreshInterval(15*time.Minute))
	ks, err := ar.Refresh(context.Background(), api.config.JwksURI)

	if err != nil {
		api.logger.Fatal("fetching jwk set failed", zap.Error(err))
	}

	config := middleware.JWTConfig{
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
			return commonAuth.GetToken(auth, ks, api.config.Audience)
		},
	}

	jwtMiddleware := middleware.JWTWithConfig(config)

	api.app.Use(middleware.RequestID())
	api.app.Use(otelecho.Middleware("pki-rest-interface", otelecho.WithSkipper(func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/docs") || strings.Contains(c.Path(), "/healthz")
	})))
	api.app.Use(logging.ZapLogger(api.logger))
	api.app.Use(middleware.Recover())
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
		c.URL = "/docs/spec.json"
	})) // default
	api.app.GET("/healthz", api.healthzHandler)
	api.app.GET("/readyz", api.readyzHandler)
	api.app.GET("/whoami", api.whoamiHandler, jwtMiddleware)

	group := api.app.Group("/ssl")
	{
		domainClient, err := domainClient(api.handlerCfg.DomainService)
		if err != nil {
			api.logger.Fatal("failed to create domain client", zap.Error(err))
		}
		sslClient, err := sslClient(api.handlerCfg.SslService)
		if err != nil {
			api.logger.Fatal("failed to create ssl client", zap.Error(err))
		}
		ssl := ssl.NewHandler(domainClient, sslClient, api.logger)
		group.Use(jwtMiddleware)
		group.GET("/", ssl.List)
		group.POST("/revoke", ssl.Revoke)
	}

	group = api.app.Group("/smime")
	{
		smimeClient, err := smimeClient(api.handlerCfg.SmimeService)
		if err != nil {
			api.logger.Fatal("failed to create smime client", zap.Error(err))
		}
		handler := smime.NewHandler(smimeClient, api.logger)
		group.Use(jwtMiddleware)
		group.GET("/", handler.List)
		group.POST("/revoke", handler.Revoke)
		group.POST("/csr", handler.HandleCsr)
	}
	ready = 1
	healthy = 1
}

func domainClient(host string) (pb.DomainServiceClient, error) {
	conn, err := api.ConnectGRPC(host)
	if err != nil {
		return nil, err
	}
	return pb.NewDomainServiceClient(conn), nil
}

func smimeClient(host string) (pb.SmimeServiceClient, error) {
	conn, err := api.ConnectGRPC(host)
	if err != nil {
		return nil, err
	}
	return pb.NewSmimeServiceClient(conn), nil
}

func sslClient(host string) (pb.SSLServiceClient, error) {
	conn, err := api.ConnectGRPC(host)
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
		if err := api.app.Start(addr); err != http.ErrServerClosed {
			api.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()
	<-stopCh
	err := api.app.Shutdown(context.Background())
	if err != nil {
		api.logger.Fatal("Stopping http server failed", zap.Error(err))
	}
}
