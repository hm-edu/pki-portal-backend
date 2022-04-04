package cfg

type SectigoConfiguration struct {
	User           string `mapstructure:"sectigo_user"`
	Password       string `mapstructure:"sectigo_password"`
	CustomerURI    string `mapstructure:"sectigo_customer"`
	SmimeProfile   int    `mapstructure:"smime_profile"`
	SmimeOrgID     int    `mapstructure:"smime_org_id"`
	SmimeTerm      int    `mapstructure:"smime_term"`
	SmimeKeyLength int    `mapstructure:"smime_key_length"`
	SmimeKeyType   string `mapstructure:"smime_key_type"`
}
