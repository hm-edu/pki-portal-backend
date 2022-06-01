package ssl

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator *model.Validator
	domain    pb.DomainServiceClient
	ssl       pb.SSLServiceClient
	tracer    trace.Tracer
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(domain pb.DomainServiceClient, ssl pb.SSLServiceClient) *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("ssl")
	return &Handler{
		validator: v,
		domain:    domain,
		ssl:       ssl,
		tracer:    tracer,
	}
}
