package ssl

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	domain    pb.DomainServiceClient
	ssl       pb.SSLServiceClient
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(domain pb.DomainServiceClient, ssl pb.SSLServiceClient) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator: v,
		domain:    domain,
		ssl:       ssl,
	}
}
