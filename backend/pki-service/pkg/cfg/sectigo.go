package cfg

import (
	"net/http"

	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/client"
	"go.uber.org/zap"
)

// SectigoConfiguration handles different configuration properties for the sectigo client
type SectigoConfiguration struct {
	User                 string `mapstructure:"sectigo_user"`
	Password             string `mapstructure:"sectigo_password"`
	CustomerURI          string `mapstructure:"sectigo_customeruri"`
	SmimeProfile         int    `mapstructure:"smime_profile"`
	SmimeProfileStandard int    `mapstructure:"smime_profile_standard"`
	SmimeOrgID           int    `mapstructure:"smime_org_id"`
	SmimeTerm            int    `mapstructure:"smime_term"`
	SmimeStudentTerm     int    `mapstructure:"smime_student_term"`
	SslProfile           int    `mapstructure:"ssl_profile"`
	SslOrgID             int    `mapstructure:"ssl_org_id"`
	SslTerm              int    `mapstructure:"ssl_term"`
	SmimeKeyLength       string `mapstructure:"smime_key_length"`
	SmimeKeyType         string `mapstructure:"smime_key_type"`
}

// CheckSectigoConfiguration checks the sectigo configuration for the syntactical correctness.
func (cfg *SectigoConfiguration) CheckSectigoConfiguration() {

	logger := zap.L()

	c := sectigo.NewClient(http.DefaultClient, zap.L(), cfg.User, cfg.Password, cfg.CustomerURI)
	profiles, err := c.ClientService.Profiles()
	if err != nil {
		logger.Fatal("fetching profiles failed", zap.Error(err))
	}
	if len(*profiles) == 0 {
		logger.Warn("no profiles found")
		return
	}
	found := helper.Any(*profiles, func(t client.ListProfileItem) bool {
		if t.ID == cfg.SmimeProfile || (t.ID == cfg.SmimeProfileStandard && cfg.SmimeProfileStandard > 0) {
			if helper.Any(t.Terms, func(i int) bool { return i == cfg.SmimeTerm }) &&
				helper.Any(t.KeyTypes[cfg.SmimeKeyType], func(i string) bool {
					return i == cfg.SmimeKeyLength
				}) {
				return true
			}
			return false
		}
		return false
	})
	if !found {
		logger.Fatal("smime profile not found")
	}

}
