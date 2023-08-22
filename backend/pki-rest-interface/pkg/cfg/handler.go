package cfg

// HandlerConfiguration holds the configuration of the different service endpoints.
type HandlerConfiguration struct {
	SmimeService   string `mapstructure:"smime_service"`
	SslService     string `mapstructure:"ssl_service"`
	DomainService  string `mapstructure:"domain_service"`
	RejectStudents bool   `mapstructure:"reject_students"`
}
