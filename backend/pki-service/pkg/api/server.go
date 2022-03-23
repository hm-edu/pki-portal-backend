//go:generate sh -c "$(go env GOPATH)/bin/swag init -g server.go --ot go --pd true "
package api

import (
	"crypto/tls"
	"net/http"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hm-edu/pki-service/pkg/api/docs"
	_ "github.com/hm-edu/pki-service/pkg/api/docs"
	commonApi "github.com/hm-edu/portal-common/api"
	commonAuth "github.com/hm-edu/portal-common/auth"
	"go.uber.org/zap"
)

var (
	healthy     int32
	ready       int32
	openApiSpec *openapi3.T
)

// @title PKI Service
// @version 2.0
// @description Go microservice for PKI management. Provides an wrapper above the sectigo API.

// @contact.name Source Code
// @contact.url  https://github.com/hm-edu/pki-service/blob/main/LICENSE

// @license.name Apache License
// @license.url https://github.com/hm-edu/pki-service/blob/main/LICENSE

// @securitydefinitions.apikey API
// @in header
// @name Authorization

type APIServer struct {
	app    *fiber.App
	logger *zap.Logger
	config *commonApi.Config
}

func NewServer(logger *zap.Logger, config *commonApi.Config) *APIServer {
	return &APIServer{app: fiber.New(fiber.Config{DisableStartupMessage: true}), logger: logger, config: config}
}

func (api *APIServer) wireRoutesAndMiddleware() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	swaggerCfg := swagger.ConfigDefault
	swaggerCfg.URL = "/docs/spec.json"

	auth := commonAuth.New(commonAuth.Config{
		Endpoint:     api.config.Endpoint,
		ClientID:     api.config.ClientID,
		ClientSecret: api.config.ClientSecret,
	})
	api.app.Use(otelfiber.Middleware("pki-service"))
	api.app.Use(recover.New())
	api.app.Use(fiberzap.New(fiberzap.Config{
		Logger: api.logger,
	}))
	api.app.Get("/docs/spec.json", func(c *fiber.Ctx) error {
		if openApiSpec == nil {
			spec, err := commonApi.ToOpenApi3(docs.SwaggerInfo)
			if err != nil {
				return err
			}
			openApiSpec = spec
		}
		return c.JSON(openApiSpec)
	})
	api.app.Get("/docs/*", swagger.New(swaggerCfg)) // default
	api.app.Get("/healthz", api.healthzHandler)
	api.app.Get("/readyz", api.readyzHandler)
	api.app.Get("/whoami", auth, api.whoamiHandler)
	api.app.Post("/acme", api.addAcmeAccount)
}

// ListenAndServe starts the http server and waits for the channel to stop the server
func (api *APIServer) ListenAndServe(stopCh <-chan struct{}) {

	api.wireRoutesAndMiddleware()
	go func() {
		addr := api.config.Host + ":" + api.config.Port
		api.logger.Info("Starting HTTP Server.", zap.String("addr", addr))
		if err := api.app.Listen(addr); err != http.ErrServerClosed {
			api.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()
	_ = <-stopCh
	err := api.app.Shutdown()
	if err != nil {
		api.logger.Error("Stopping http server failed", zap.Error(err))
	}
}
