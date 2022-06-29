package cfg

// SectigoConfiguration handles different configuration properties for the sectigo client
type SectigoConfiguration struct {
	User        string `mapstructure:"sectigo_user"`
	Password    string `mapstructure:"sectigo_password"`
	CustomerURI string `mapstructure:"sectigo_customeruri"`
}
