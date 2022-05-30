package eab

import (
	"github.com/hm-edu/portal-common/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator     *model.Validator
	tracer        trace.Tracer
	provisionerID string
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(ProvisionerID string) *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("eab")
	return &Handler{
		validator:     v,
		tracer:        tracer,
		provisionerID: ProvisionerID,
	}
}
