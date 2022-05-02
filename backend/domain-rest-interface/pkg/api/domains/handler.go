package domains

import (
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	"github.com/hm-edu/portal-common/model"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	domainStore *store.DomainStore
	validator   *model.Validator
	logger      *zap.Logger
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(ds *store.DomainStore) *Handler {
	v := model.NewValidator()
	return &Handler{
		domainStore: ds,
		validator:   v,
		logger:      zap.L(),
	}
}
