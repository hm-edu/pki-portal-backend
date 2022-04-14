//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/brpaz/echozap"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/hm-edu/pki-rest-interface/pkg/api/docs"
	"github.com/hm-edu/pki-rest-interface/pkg/api/smime"
	"github.com/hm-edu/pki-rest-interface/pkg/api/ssl"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	commonApi "github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lestrrat-go/jwx/jwk"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	echoSwagger "github.com/swaggo/echo-swagger"
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
	sectigoCfg *cfg.SectigoConfiguration
}

// NewServer creates a new server
func NewServer(logger *zap.Logger, config *commonApi.Config, sectigoCfg *cfg.SectigoConfiguration) *Server {
	return &Server{app: echo.New(), logger: logger, config: config, sectigoCfg: sectigoCfg}
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
	api.app.Use(otelecho.Middleware("pki-rest-interface"))
	api.app.Use(middleware.Recover())
	api.app.Use(echozap.ZapLogger(api.logger))
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

	// According to https://go.dev/src/net/http/client.go:
	// "Clients are safe for concurrent use by multiple goroutines."
	// => one http client is fine ;)
	c := sectigo.NewClient(http.DefaultClient, api.logger, api.sectigoCfg.User, api.sectigoCfg.Password, api.sectigoCfg.CustomerURI)

	group := api.app.Group("/ssl")
	{
		ssl := ssl.NewHandler(c, api.sectigoCfg)
		group.Use(jwtMiddleware)
		group.GET("/", ssl.List)
		group.POST("/revoke", ssl.Revoke)
	}

	group = api.app.Group("/smime")
	{
		handler := smime.NewHandler(c, api.sectigoCfg)
		group.Use(jwtMiddleware)
		group.GET("/", handler.List)
		group.POST("/revoke", handler.Revoke)
		group.POST("/csr", handler.HandleCsr)
	}

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
	_ = <-stopCh
	err := api.app.Shutdown(context.Background())
	if err != nil {
		api.logger.Fatal("Stopping http server failed", zap.Error(err))
	}
}
