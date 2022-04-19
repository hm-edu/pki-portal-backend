package smime

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	smime     pb.SmimeServiceClient
	logger    *zap.Logger
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(smime pb.SmimeServiceClient, logger *zap.Logger) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator: v,
		logger:    logger,
		smime:     smime,
	}
}
