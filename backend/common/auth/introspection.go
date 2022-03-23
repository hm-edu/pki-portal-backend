package auth

// IntrospectionResult represents the result of the introspection request.
type IntrospectionResult struct {
	// Active is a boolean indicator of whether or not the presented token is currently active.  The specifics of a token's "active" state will vary depending on the implementation of the authorization server and the information it keeps about its tokens, but a "true" value return for the "active" property will generally indicate that a given token has been issued by this authorization server, has not been revoked by the resource owner, and is within its given time window of validity (e.g., after its issuance time and before its expiration time).
	//
	// required: true
	Active bool `json:"active"`

	// Scope is a JSON string containing a space-separated list of scopes associated with this token.
	Scope string `json:"scope,omitempty"`

	// ID is aclient identifier for the OAuth 2.0 client that requested this token.
	ClientID string `json:"client_id"`

	// Subject of the token, as defined in JWT [RFC7519]. Usually a machine-readable identifier of the resource owner who authorized this token.
	Subject string `json:"sub"`

	// Expires at is an unix timestamp, indicating when this token will expire.
	ExpiresAt int64 `json:"exp"`

	// Issued at is an integer timestamp, indicating when this token was originally issued.
	IssuedAt int64 `json:"iat"`

	// NotBefore is an integer timestamp, indicating when this token is not to be used before.
	NotBefore int64 `json:"nbf"`

	// Username is a human-readable identifier for the resource owner who authorized this token.
	Username string `json:"username,omitempty"`

	// Audience contains a list of the token's intended audiences.
	Audience []string `json:"aud"`

	// IssuerURL is a string representing the issuer of this token
	Issuer string `json:"iss"`

	// TokenType is the introspected token's type, typically `Bearer`.
	TokenType string `json:"token_type"`

	// Extra is optional data set by the session.
	Extra map[string]interface{} `json:"ext,omitempty"`
}
