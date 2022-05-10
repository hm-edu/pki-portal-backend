package eab

import (
	"github.com/hm-edu/portal-common/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	logger    *zap.Logger
	tracer    trace.Tracer
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler() *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("eab")
	return &Handler{
		validator: v,
		logger:    zap.L(),
		tracer:    tracer,
	}
}
