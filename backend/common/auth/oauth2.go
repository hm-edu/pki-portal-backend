package auth

import (
	"errors"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

// UserFromRequest extracts username (or email) from the previously claimset.
func UserFromRequest(c echo.Context) string {
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(jwt.MapClaims)["email"].(string)
	return user
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
