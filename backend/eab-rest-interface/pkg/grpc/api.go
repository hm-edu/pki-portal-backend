package grpc

import (
	"context"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/hm-edu/eab-rest-interface/ent"
	"github.com/hm-edu/eab-rest-interface/ent/eabkey"
	"github.com/hm-edu/eab-rest-interface/pkg/database"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

type eabAPIServer struct {
	pb.UnimplementedEABServiceServer
	domainService pb.DomainServiceClient
	logger        *zap.Logger
	provisionerID string
}

func newEabAPIServer(domainService pb.DomainServiceClient, logger *zap.Logger, provisionerID string) *eabAPIServer {
	return &eabAPIServer{domainService: domainService, logger: logger, provisionerID: provisionerID}
}

// ResolveAccountId resolves the user using the provided account id and returns the EAB ID and the username.
func (api *eabAPIServer) ResolveAccountId(ctx context.Context, req *pb.ResolveAccountIdRequest) (*pb.ResolveAccountIdResponse, error) { // nolint
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := api.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}
	log = log.With(zap.String("user", req.AccountId))

	eak, err := database.DB.NoSQL.GetExternalAccountKeyByAccountID(ctx, api.provisionerID, req.AccountId)
	if err != nil {
		hub.CaptureException(err)
		log.Error("Error looking up user", zap.String("acc", req.AccountId), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}
	key, err := database.DB.Db.EABKey.Query().Where(eabkey.EabKey(eak.ID)).First(ctx)

	if hub != nil && hub.Scope() != nil {
		hub.Scope().SetUser(sentry.User{Email: key.User})
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			log.Warn("Key not found", zap.String("key", eak.ID), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "Key not found")
		}
		hub.CaptureException(err)
		log.Error("Error looking up user", zap.String("key", eak.ID), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}
	log.Info("User found", zap.String("key", eak.ID), zap.String("user", key.User))
	return &pb.ResolveAccountIdResponse{User: key.User, EabKey: eak.ID}, nil
}

// CheckEABPermissions resolves the user and validates the issue permission for the requested domains.
func (api *eabAPIServer) CheckEABPermissions(ctx context.Context, req *pb.CheckEABPermissionRequest) (*pb.CheckEABPermissionResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := api.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}
	log = log.With(zap.String("user", req.EabKey))

	key, err := database.DB.Db.EABKey.Query().Where(eabkey.EabKey(req.EabKey)).First(ctx)

	if hub != nil && hub.Scope() != nil {
		hub.Scope().SetUser(sentry.User{Email: key.User})
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			log.Warn("Key not found", zap.String("key", req.EabKey), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "Key not found")
		}
		hub.CaptureException(err)
		log.Error("Error looking up user", zap.String("key", req.EabKey), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Checking permissions", Category: "info"}, nil)
	log.Info("Checking registrations for user", zap.String("key", key.EabKey), zap.String("user", key.User), zap.Strings("domains", req.Domains))
	permissions, err := api.domainService.CheckPermission(ctx, &pb.CheckPermissionRequest{User: key.User, Domains: req.Domains})
	if err != nil {
		hub.CaptureException(err)
		log.Error("Error checking permissions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error checking permissions")
	}
	missing := helper.Map(helper.Where(permissions.Permissions, func(t *pb.Permission) bool { return !t.Granted }), func(t *pb.Permission) string { return t.Domain })
	log.Info("Permissions checked", zap.Strings("missing", missing), zap.String("key", key.EabKey), zap.String("user", key.User))
	return &pb.CheckEABPermissionResponse{Missing: missing}, nil
}
