package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestExtractScopesWithScopeSpaceSeparated(t *testing.T) {
	claims := jwt.MapClaims{
		"scope": "openid email",
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}

func TestExtractScopesWithScopesSpaceSeparated(t *testing.T) {
	claims := jwt.MapClaims{
		"scopes": "openid email",
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}

func TestExtractScopesWithScpSpaceSeparated(t *testing.T) {
	claims := jwt.MapClaims{
		"scp": "openid email",
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}

func TestExtractScopesWithScopeArray(t *testing.T) {
	claims := jwt.MapClaims{
		"scope": []string{"openid", "email"},
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}

func TestExtractScopesWithScopesArray(t *testing.T) {
	claims := jwt.MapClaims{
		"scopes": []string{"openid", "email"},
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}

func TestExtractScopesWithScpArray(t *testing.T) {
	claims := jwt.MapClaims{
		"scp": []string{"openid", "email"},
	}

	scopes, ok := ExtractScopes(claims)
	assert.True(t, ok)
	assert.Contains(t, scopes, "openid")
}
