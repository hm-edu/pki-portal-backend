package auth

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ExtractScopes extracts the scope from a given claim set
// and returns them in normalized form
func ExtractScopes(claims jwt.MapClaims) ([]string, bool) {
	claim, ok := extractClaim(claims, []string{"scope", "scopes", "scp"})
	if !ok {
		return nil, false
	}

	if scope, ok := claim.(string); ok {
		return strings.Split(scope, " "), true
	}

	if scopes, ok := claim.([]string); ok {
		return scopes, true
	}

	return nil, false
}

// extractClaim attempts to extract a claim from a given set of claims
// by returning the first existing claim from supportedKeys
func extractClaim(claims jwt.MapClaims, supportedKeys []string) (interface{}, bool) {
	for _, key := range supportedKeys {
		if claim, ok := claims[key]; ok {
			return claim, true
		}
	}

	return nil, false
}
