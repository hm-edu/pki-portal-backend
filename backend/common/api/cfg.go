package api

// Config of the REST-API
type Config struct {
	Host               string   `mapstructure:"host"`
	Port               string   `mapstructure:"port"`
	Audience           string   `mapstructure:"audience"`
	JwksURI            string   `mapstructure:"jwks_uri"`
	CorsAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
	SentryDSN          string   `mapstructure:"sentry_dsn"`
}
