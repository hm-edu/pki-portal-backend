package ssl

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	domain    pb.DomainServiceClient
	ssl       pb.SSLServiceClient
	logger    *zap.Logger
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(domain pb.DomainServiceClient, ssl pb.SSLServiceClient, logger *zap.Logger) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator: v,
		domain:    domain,
		ssl:       ssl,
		logger:    logger,
	}
}
