package cfg

// PKIConfiguration handles different configuration properties for the sectigo client
type PKIConfiguration struct {
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	TotpSeed           string `mapstructure:"totp_seed"`
	ValidationUser     string `mapstructure:"validation_user"`
	ValidationPassword string `mapstructure:"validation_password"`
	ValidationTotpSeed string `mapstructure:"validation_totp_seed"`
	SmimeKeyLength     string `mapstructure:"smime_key_length"`
	CertType           string `mapstructure:"cert_type"`

	// SslCa selects the CA used for issuing server certificates
	// ("harica" or "letsencrypt").
	SslCa string `mapstructure:"ssl_ca"`
	// AcmeEmail is the contact mail address of the ACME account.
	AcmeEmail string `mapstructure:"acme_email"`
	// AcmeDirectory is the directory URL of the ACME CA.
	AcmeDirectory string `mapstructure:"acme_directory"`
	// AcmeAccountKey is the path to the PEM encoded ACME account key.
	// The key is generated (and the account registered) on first start.
	AcmeAccountKey string `mapstructure:"acme_account_key"`
	// AcmeDNSConfig is the path to the YAML file mapping DNS zones to the
	// TSIG keys used for the DNS-01 validation.
	AcmeDNSConfig string `mapstructure:"acme_dns_config"`
}
