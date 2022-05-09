package smime

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	smime     pb.SmimeServiceClient
	logger    *zap.Logger
	tracer    trace.Tracer
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(smime pb.SmimeServiceClient, logger *zap.Logger) *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("smime")
	return &Handler{
		validator: v,
		logger:    logger,
		smime:     smime,
		tracer:    tracer,
	}
}
