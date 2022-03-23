package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hm-edu/portal-common/models"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("oauth2-introspection")

// Config of the OAuth2-Introspection middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid token.
	// Optional. Default: nil
	SuccessHandler fiber.Handler

	// ErrorHandler to use in case of a failing token introspection
	ErrorHandler models.ErrorHandler

	// Endpoint to use for the token introspection
	Endpoint string

	// ClientID to use for the client authentication
	ClientID string

	// ClientSecret to use for the client authentication
	ClientSecret string
}

func makeCfg(config []Config) (cfg Config) {
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.SuccessHandler == nil {
		cfg.SuccessHandler = func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err models.Error) error {
			return c.Status(err.Status).JSON(err)
		}
	}
	return cfg
}

// OAuthIntrospectionHandler is the actual handler that does the introspection
func OAuthIntrospectionHandler(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		_, span := tracer.Start(ctx, "checkAuthentication")
		defer span.End()
		auth := c.Get("Authorization")
		if auth == "" {
			span.RecordError(errors.New("missing authorization"))
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusUnauthorized, Message: "Unauthorized", Details: "Missing or malformed access token"})
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			span.RecordError(errors.New("missing authorization"))
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusUnauthorized, Message: "Unauthorized", Details: "Missing or malformed access token"})
		}
		resp, err := http.PostForm(cfg.Endpoint, url.Values{"client_id": {cfg.ClientID}, "client_secret": {cfg.ClientSecret}, "token": {token}})
		if err != nil {
			span.RecordError(err)
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusInternalServerError, Message: "Internal Server Error", Details: "Introspection Failed"})
		}
		if resp.StatusCode != 200 {
			span.RecordError(fmt.Errorf("server returned %v", resp.StatusCode))
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusInternalServerError, Message: "Internal Server Error", Details: "Introspection Failed"})
		}
		defer func(i io.ReadCloser) { _ = i.Close() }(resp.Body)

		obj := new(IntrospectionResult)
		err = json.NewDecoder(resp.Body).Decode(obj)

		if err != nil {
			span.RecordError(err)
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusInternalServerError, Message: "Internal Server Error", Details: "Introspection Failed"})
		}

		if !obj.Active {
			span.AddEvent("Inactive token")
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusForbidden, Message: "Forbidden", Details: "Token is inactive"})
		}
		c.Locals("scopes", strings.Split(obj.Scope, " "))
		c.Locals("subject", obj.Subject)
		c.Locals("username", obj.Username)
		span.AddEvent("Token valid")
		return cfg.SuccessHandler(c)
	}
}

// New generates a new handler that shal be used as a middleware.
func New(config ...Config) fiber.Handler {
	cfg := makeCfg(config)
	handler := OAuthIntrospectionHandler(cfg)
	// Return middleware handler
	return func(c *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}
		return handler(c)
	}
}
