package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// config is used to configure the mux middleware.
type config struct {
	Skipper middleware.Skipper
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithSkipper specifies a skipper for allowing requests to skip generating spans.
func WithSkipper(skipper middleware.Skipper) Option {
	return optionFunc(func(cfg *config) {
		cfg.Skipper = skipper
	})
}

// ZapLogger is a middleware and zap to provide an "access log" like logging for each request.
func ZapLogger(log *zap.Logger, opts ...Option) echo.MiddlewareFunc {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			req := c.Request()

			hub := sentryecho.GetHubFromContext(c)
			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id != "" {
				fields = append(fields, zap.String("request_id", id))
			}
			fields = append(fields, zapsentry.NewScopeFromScope(hub.Scope()))
			logger := log.With(fields...)

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), LoggingContextKey, logger)))
			if cfg.Skipper(c) {
				return next(c)
			}
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			res := c.Response()

			fields = append(fields,
				zap.String("latency", time.Since(start).String()),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
			)

			n := res.Status
			switch {
			case n >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}
