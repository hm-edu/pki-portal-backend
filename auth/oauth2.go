package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hm-edu/portal-common/models"
)

type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid token.
	// Optional. Default: nil
	SuccessHandler fiber.Handler

	ErrorHandler models.ErrorHandler

	Endpoint string

	ClientID string

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
func OAuthIntrospectionHandler(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusUnauthorized, Message: "Unauthorized", Details: "Missing or malformed access token"})
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusUnauthorized, Message: "Unauthorized", Details: "Missing or malformed access token"})
		}
		resp, err := http.PostForm(cfg.Endpoint, url.Values{"client_id": {cfg.ClientID}, "client_secret": {cfg.ClientSecret}, "token": {token}})
		if err != nil {
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusInternalServerError, Message: "Internal Server Error", Details: "Introspection Failed"})
		}
		defer resp.Body.Close()
		obj := new(IntrospectionResult)
		err = json.NewDecoder(resp.Body).Decode(obj)

		if err != nil {
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusInternalServerError, Message: "Internal Server Error", Details: "Introspection Failed"})
		}

		if !obj.Active {
			return cfg.ErrorHandler(c, models.Error{Status: fiber.StatusForbidden, Message: "Forbidden", Details: "Token is inactive"})
		}
		c.Locals("scopes", strings.Split(obj.Scope, " "))
		c.Locals("subject", obj.Subject)
		c.Locals("username", obj.Username)
		return cfg.SuccessHandler(c)
	}
}
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
