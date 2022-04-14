package smime

import (
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	"github.com/hm-edu/portal-common/model"
	"github.com/hm-edu/sectigo-client/sectigo"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	sectigoClient *sectigo.Client
	sectigoCfg    *cfg.SectigoConfiguration
	validator     *model.Validator
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(c *sectigo.Client, sectigoCfg *cfg.SectigoConfiguration) *Handler {
	v := model.NewValidator()
	return &Handler{
		sectigoClient: c,
		validator:     v,
		sectigoCfg:    sectigoCfg,
	}
}
