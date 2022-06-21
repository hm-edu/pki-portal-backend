package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwk"
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
			return user["email"].(string), nil
		}
	}

	return "", errors.New("unable to extract user from request")
}

// GetToken extracts the JWT Token from header and also performs a check on the passed audience.
func GetToken(auth string, ks jwk.Set, aud string) (interface{}, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return GetKey(t, ks)
	}

	// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
	token, err := jwt.Parse(auth, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if !token.Claims.(jwt.MapClaims).VerifyAudience(aud, true) {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

// GetKey looks for a signature key in the passed JWK Set that matches to the key id in the passed header.
func GetKey(token *jwt.Token, ks jwk.Set) (interface{}, error) {
	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have a key ID in the kid field")
	}

	key, found := ks.LookupKeyID(keyID)

	if !found {
		return nil, fmt.Errorf("unable to find key %q", keyID)
	}

	var pubkey interface{}
	if err := key.Raw(&pubkey); err != nil {
		return nil, fmt.Errorf("Unable to get the public key. Error: %s", err.Error())
	}

	return pubkey, nil
}

// HasScope tries to extract the scope from the provided token.
func HasScope(scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if token, ok := c.Get("user").(*jwt.Token); ok {
				if user, ok := token.Claims.(jwt.MapClaims); ok {
					if scp, ok := user["scp"].(string); ok {
						if !helper.Contains(strings.Split(scp, " "), scope) {
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
