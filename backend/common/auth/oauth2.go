package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
)

// SubFromRequest extracts username (or email) from the previously claimset.
func SubFromRequest(token interface{}) (string, error) {
	if token, ok := token.(*jwt.Token); ok {
		if user, ok := token.Claims.(jwt.MapClaims); ok {
			return user["sub"].(string), nil
		}
	}

	return "", errors.New("unable to extract user from request")
}

// UserFromRequest extracts username (or email) from the previously claimset.
func UserFromRequest(c echo.Context) (string, error) {
	if token, ok := c.Get("user").(*jwt.Token); ok {
		if user, ok := token.Claims.(jwt.MapClaims); ok {
			if _, ok := user["email"]; !ok {
				return "", errors.New("email not found in token")
			}
			return user["email"].(string), nil
		}
	}

	return "", errors.New("unable to extract user from request")
}

// GetToken extracts the JWT Token from header and also performs a check on the passed audience.
func GetToken(auth string, keyfunc keyfunc.Keyfunc, aud string) (interface{}, error) {
	// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
	token, err := jwt.Parse(auth, keyfunc.Keyfunc, jwt.WithExpirationRequired(), jwt.WithAudience(aud))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// HasScope tries to extract the scope from the provided token.
func HasScope(scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if token, ok := c.Get("user").(*jwt.Token); ok {
				if user, ok := token.Claims.(jwt.MapClaims); ok {
					if scopes, ok := ExtractScopes(user); ok {
						if !helper.Contains(scopes, scope) {
							return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("scope %s missing", scope))
						}
						return next(c)
					}
				}
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "extracting scope failed")
		}
	}
}
