package model

import "strings"

// NormalizeUser canonicalizes a user identifier (an e-mail address) by stripping
// surrounding whitespace and lowercasing it. Owners, delegations and the identity
// taken from the JWT are all passed through this, so they compare equal no matter
// how the address was typed or which case the identity provider returns.
func NormalizeUser(user string) string {
	return strings.ToLower(strings.TrimSpace(user))
}
