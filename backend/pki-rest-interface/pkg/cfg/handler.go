package cfg

type HandlerConfiguration struct {
	SmimeService  string `mapstructure:"smime_service"`
	SslService    string `mapstructure:"ssl_service"`
	DomainService string `mapstructure:"domain_service"`
}
