package grpc

import (
	"context"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"

	"go.uber.org/zap"
)

type domainAPIServer struct {
	pb.UnimplementedDomainServiceServer
	store  *store.DomainStore
	logger *zap.Logger
	admins []string
}

func newDomainAPIServer(store *store.DomainStore, logger *zap.Logger, admins []string) *domainAPIServer {
	return &domainAPIServer{store: store, logger: logger, admins: admins}
}

func (api *domainAPIServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	span := sentry.StartSpan(ctx, "Check Permission")
	defer span.Finish()
	ctx = span.Context()
	log := api.logger
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}
	domains, err := api.store.ListDomains(ctx, req.User, true, false)
	if err != nil {
		return nil, err
	}
	log.Info("Checking permissions", zap.String("user", req.User), zap.Strings("domains", req.Domains))
	permissions := helper.Map(req.Domains, func(t string) *pb.Permission {
		if helper.Any(domains, func(d *ent.Domain) bool { return d.Fqdn == t }) {
			log.Info("Permission granted", zap.String("user", req.User), zap.String("domain", t))
			return &pb.Permission{Domain: t, Granted: true}
		}
		log.Info("Permission denied", zap.String("user", req.User), zap.String("domain", t))
		return &pb.Permission{Domain: t, Granted: false}
	})
	log.Info("Checked permissions", zap.String("user", req.User), zap.Any("permissions", permissions))
	resp := pb.CheckPermissionResponse{Permissions: permissions}

	return &resp, nil
}

func (api *domainAPIServer) CheckRegistration(ctx context.Context, req *pb.CheckRegistrationRequest) (*pb.CheckRegistrationResponse, error) {

	span := sentry.StartSpan(ctx, "Check Registration")
	defer span.Finish()
	ctx = span.Context()
	log := api.logger
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	log.Info("Checking registrations domains", zap.Strings("domains", req.Domains))
	domains, err := api.store.ListAllDomains(ctx, true)
	if err != nil {
		log.Error("Checking registrations failed", zap.Strings("domains", req.Domains), zap.Error(err))
		return nil, err
	}

	missing := helper.Where(req.Domains, func(t string) bool {
		return !helper.Any(domains, func(d string) bool {
			return d == t
		})
	})
	log.Info("Checked registrations", zap.Strings("domains", req.Domains), zap.Strings("missing", missing))
	return &pb.CheckRegistrationResponse{Missing: missing}, nil
}

func (api *domainAPIServer) ListDomains(ctx context.Context, req *pb.ListDomainsRequest) (*pb.ListDomainsResponse, error) {
	span := sentry.StartSpan(ctx, "List Domains")
	defer span.Finish()
	ctx = span.Context()
	log := api.logger
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	log.Info("Listing domains", zap.String("user", req.User))
	domains, err := api.store.ListDomains(ctx, req.User, req.Approved, helper.Contains(api.admins, req.User))
	if err != nil {
		log.Error("Listing domains failed", zap.String("user", req.User), zap.Error(err))
		return nil, err
	}
	log.Debug("Listed domains", zap.String("user", req.User), zap.Any("domains", domains))
	resp := pb.ListDomainsResponse{Domains: helper.Map(domains, func(t *ent.Domain) string { return t.Fqdn })}
	return &resp, nil
}
