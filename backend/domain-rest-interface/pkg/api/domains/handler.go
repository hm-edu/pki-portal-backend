package domains

import (
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	domainStore *store.DomainStore
	pkiService  pb.SSLServiceClient
	validator   *model.Validator
	admins      []string
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(ds *store.DomainStore, pkiSerivce pb.SSLServiceClient, admins []string) *Handler {
	v := model.NewValidator()
	return &Handler{
		domainStore: ds,
		validator:   v,
		pkiService:  pkiSerivce,
		admins:      admins,
	}
}
