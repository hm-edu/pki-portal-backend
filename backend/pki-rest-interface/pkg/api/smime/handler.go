package smime

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	ssl       pb.SSLServiceClient
	logger    *zap.Logger
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(pb.SmimeServiceClient, *zap.Logger) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator: v,
	}
}
