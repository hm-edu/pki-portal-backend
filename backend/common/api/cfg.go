package api

type Config struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	ClientID     string `mapstructure:"oauth2_id"`
	ClientSecret string `mapstructure:"oauth2_secret"`
	Endpoint     string `mapstructure:"oauth2_endpoint"`
}
