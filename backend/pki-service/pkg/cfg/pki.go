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
}
