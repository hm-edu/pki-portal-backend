package smime

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	smime     pb.SmimeServiceClient
	tracer    trace.Tracer
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(smime pb.SmimeServiceClient) *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("smime")
	return &Handler{
		validator: v,
		smime:     smime,
		tracer:    tracer,
	}
}
