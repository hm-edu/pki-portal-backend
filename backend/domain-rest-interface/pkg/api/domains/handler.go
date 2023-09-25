package domains

import (
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	domainStore *store.DomainStore
	pkiService  pb.SSLServiceClient
	validator   *model.Validator
	tracer      trace.Tracer
	admins      []string
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(ds *store.DomainStore, pkiSerivce pb.SSLServiceClient, admins []string) *Handler {
	v := model.NewValidator()
	tracer := otel.GetTracerProvider().Tracer("domains")
	return &Handler{
		domainStore: ds,
		validator:   v,
		tracer:      tracer,
		pkiService:  pkiSerivce,
		admins:      admins,
	}
}
