package eab

import (
	"github.com/hm-edu/portal-common/model"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator     *model.Validator
	provisionerID string
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(ProvisionerID string) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator:     v,
		provisionerID: ProvisionerID,
	}
}
