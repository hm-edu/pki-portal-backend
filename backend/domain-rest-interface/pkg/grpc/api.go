package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type domainAPIServer struct {
	pb.UnimplementedDomainServiceServer
	store  *store.DomainStore
	logger *zap.Logger
	tracer trace.Tracer
}

func newDomainAPIServer(store *store.DomainStore, logger *zap.Logger) *domainAPIServer {

	tracer := otel.GetTracerProvider().Tracer("domains")
	return &domainAPIServer{store: store, logger: logger, tracer: tracer}
}

func (api *domainAPIServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {

	domains, err := api.store.ListDomains(ctx, req.User, true)
	if err != nil {
		return nil, err
	}

	permissions := helper.Map(req.Domains, func(t string) *pb.Permission {
		if helper.Any(domains, func(d *ent.Domain) bool {
			return d.Fqdn == t || strings.HasSuffix(t, fmt.Sprintf(".%s", d.Fqdn))
		}) {
			return &pb.Permission{Domain: t, Granted: true}
		}
		return &pb.Permission{Domain: t, Granted: false}
	})

	resp := pb.CheckPermissionResponse{Permissions: permissions}

	return &resp, nil
}

func (api *domainAPIServer) CheckRegistration(ctx context.Context, req *pb.CheckRegistrationRequest) (*pb.CheckRegistrationResponse, error) {

	ctx, span := api.tracer.Start(ctx, "CheckRegistration")
	defer span.End()

	domains, err := api.store.ListAllDomains(ctx, true)
	if err != nil {
		span.RecordError(err)
		api.logger.Error("Checking registrations failed", zap.Strings("domains", req.Domains), zap.Error(err))
		return nil, err
	}

	missing := helper.Where(req.Domains, func(t string) bool {
		return !helper.Any(domains, func(d string) bool {
			return d == t
		})
	})
	api.logger.Info("Checking registration", zap.Strings("domains", req.Domains), zap.Strings("missing", missing))
	return &pb.CheckRegistrationResponse{Missing: missing}, nil
}

func (api *domainAPIServer) ListDomains(ctx context.Context, req *pb.ListDomainsRequest) (*pb.ListDomainsResponse, error) {
	ctx, span := api.tracer.Start(ctx, "ListDomains")
	defer span.End()

	domains, err := api.store.ListDomains(ctx, req.User, req.Approved)
	if err != nil {
		span.RecordError(err)
		api.logger.Error("Listing domains failed", zap.String("user", req.User), zap.Error(err))
		return nil, err
	}

	resp := pb.ListDomainsResponse{Domains: helper.Map(domains, func(t *ent.Domain) string { return t.Fqdn })}
	return &resp, nil
}
