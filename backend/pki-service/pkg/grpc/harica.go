package grpc

import (
	"context"
	"errors"
	"net/http"
	"time"

	harica "github.com/hm-edu/harica/client"
	"github.com/hm-edu/pki-service/pkg/cfg"

	"go.uber.org/zap"
)

const (
	// haricaRefreshInterval is the safety margin for session reuse: a login is
	// only performed if the current token expires within this interval.
	haricaRefreshInterval = 10 * time.Minute
	// haricaRequestTimeout bounds a single HTTP request so that hanging
	// connections surface as errors and can be retried.
	haricaRequestTimeout = 2 * time.Minute

	haricaMaxAttempts    = 4
	haricaInitialBackoff = 2 * time.Second
)

// haricaClients bundles the shared HARICA API clients. Both clients are
// created once at startup and reused for all requests. Sessions are kept
// alive as long as the token is valid; a fresh login only happens close to
// the token expiry (or forced after an authorization failure).
type haricaClients struct {
	// client is bound to the regular account and used for requesting certificates.
	client *harica.Client
	// validation is bound to the validation account and used for approving
	// requests and revoking certificates.
	validation *harica.Client
}

func newHaricaClients(cfg *cfg.PKIConfiguration) (*haricaClients, error) {
	client, err := harica.NewClient(
		cfg.User,
		cfg.Password,
		cfg.TotpSeed,
		harica.WithRetry(3),
		harica.WithRefreshInterval(haricaRefreshInterval),
		harica.WithRequestTimeout(haricaRequestTimeout),
	)
	if err != nil {
		return nil, err
	}
	validation, err := harica.NewClient(
		cfg.ValidationUser,
		cfg.ValidationPassword,
		cfg.ValidationTotpSeed,
		harica.WithRetry(3),
		harica.WithRefreshInterval(haricaRefreshInterval),
		harica.WithRequestTimeout(haricaRequestTimeout),
	)
	if err != nil {
		return nil, err
	}
	return &haricaClients{client: client, validation: validation}, nil
}

func isAuthError(err error) bool {
	var codeErr *harica.UnexpectedResponseCodeError
	if errors.As(err, &codeErr) {
		return codeErr.Code == http.StatusUnauthorized || codeErr.Code == http.StatusForbidden
	}
	return false
}

func isRetryableError(err error) bool {
	var codeErr *harica.UnexpectedResponseCodeError
	if errors.As(err, &codeErr) {
		switch {
		case codeErr.Code >= http.StatusInternalServerError:
			return true
		case codeErr.Code == http.StatusRequestTimeout, codeErr.Code == http.StatusTooManyRequests:
			return true
		default:
			// Auth errors are retried after a forced re-login, all other
			// client errors are permanent.
			return isAuthError(err)
		}
	}
	// Anything else (network errors, timeouts, HTML error pages, ...) is
	// considered transient.
	return true
}

// retryHarica runs fn with exponential backoff. Before each attempt the
// session is refreshed lazily, i.e. the existing token is reused unless it is
// (about to be) expired. If HARICA rejects the token anyway, a fresh login is
// forced before the next attempt.
func retryHarica[T any](ctx context.Context, logger *zap.Logger, client *harica.Client, op string, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error
	backoff := haricaInitialBackoff
	for attempt := 1; attempt <= haricaMaxAttempts; attempt++ {
		if attempt > 1 {
			logger.Warn("Retrying HARICA request",
				zap.String("operation", op),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.Error(lastErr))
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}
		if err := client.SessionRefresh(false); err != nil {
			lastErr = err
			continue
		}
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableError(err) {
			return zero, err
		}
		if isAuthError(err) {
			if err := client.SessionRefresh(true); err != nil {
				lastErr = err
			}
		}
	}
	return zero, lastErr
}

// retryHaricaVoid is retryHarica for operations without a result.
func retryHaricaVoid(ctx context.Context, logger *zap.Logger, client *harica.Client, op string, fn func() error) error {
	_, err := retryHarica(ctx, logger, client, op, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

// runHaricaOnce ensures a valid session and runs fn exactly once. It is used
// for non-idempotent operations where a retry could create duplicates.
func runHaricaOnce[T any](client *harica.Client, fn func() (T, error)) (T, error) {
	if err := client.SessionRefresh(false); err != nil {
		var zero T
		return zero, err
	}
	return fn()
}
