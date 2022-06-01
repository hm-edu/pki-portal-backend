package grpc

import (
	"context"

	"github.com/hm-edu/eab-rest-interface/ent"
	"github.com/hm-edu/eab-rest-interface/ent/eabkey"
	"github.com/hm-edu/eab-rest-interface/pkg/database"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type eabAPIServer struct {
	pb.UnimplementedEABServiceServer
	domainService pb.DomainServiceClient
	logger        *zap.Logger
	tracer        trace.Tracer
	provisionerID string
}

func newEabAPIServer(domainService pb.DomainServiceClient, logger *zap.Logger, provisionerID string) *eabAPIServer {
	tracer := otel.GetTracerProvider().Tracer("eab")
	return &eabAPIServer{domainService: domainService, logger: logger, tracer: tracer, provisionerID: provisionerID}
}

// ResolveAccountId resolves the user using the provided account id and returns the EAB ID and the username.
func (api *eabAPIServer) ResolveAccountId(ctx context.Context, req *pb.ResolveAccountIdRequest) (*pb.ResolveAccountIdResponse, error) { // nolint
	ctx, span := api.tracer.Start(ctx, "CheckEABPermissions")
	defer span.End()
	eak, err := database.DB.NoSQL.GetExternalAccountKeyByAccountID(ctx, api.provisionerID, req.AccountId)
	if err != nil {
		span.RecordError(err)
		span.AddEvent("Error looking up user")
		api.logger.Error("Error looking up user", zap.String("acc", req.AccountId), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}
	key, err := database.DB.Db.EABKey.Query().Where(eabkey.EabKey(eak.ID)).First(ctx)

	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			span.AddEvent("Key not found")
			api.logger.Warn("Key not found", zap.String("key", eak.ID), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "Key not found")
		}
		span.RecordError(err)
		span.AddEvent("Error looking up user")
		api.logger.Error("Error looking up user", zap.String("key", eak.ID), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}
	return &pb.ResolveAccountIdResponse{User: key.User, EabKey: eak.ID}, nil
}

// CheckEABPermissions resolves the user and validates the issue permission for the requested domains.
func (api *eabAPIServer) CheckEABPermissions(ctx context.Context, req *pb.CheckEABPermissionRequest) (*pb.CheckEABPermissionResponse, error) {

	ctx, span := api.tracer.Start(ctx, "CheckEABPermissions")
	defer span.End()
	span.SetAttributes(attribute.String("key", req.EabKey), attribute.StringSlice("domains", req.Domains))
	key, err := database.DB.Db.EABKey.Query().Where(eabkey.EabKey(req.EabKey)).First(ctx)

	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			span.AddEvent("Key not found")
			api.logger.Warn("Key not found", zap.String("key", req.EabKey), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "Key not found")
		}
		span.RecordError(err)
		span.AddEvent("Error looking up user")
		api.logger.Error("Error looking up user", zap.String("key", req.EabKey), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error looking up user")
	}

	span.SetAttributes(attribute.String("key", req.EabKey), attribute.StringSlice("domains", req.Domains), attribute.String("user", key.User))
	api.logger.Info("Checking registrations for user", zap.String("key", key.EabKey), zap.String("user", key.User), zap.Strings("domains", req.Domains))
	permissions, err := api.domainService.CheckPermission(ctx, &pb.CheckPermissionRequest{User: key.User, Domains: req.Domains})
	if err != nil {
		span.RecordError(err)
		span.AddEvent("Error checking permissions")
		api.logger.Error("Error checking permissions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error checking permissions")
	}

	missing := helper.Map(helper.Where(permissions.Permissions, func(t *pb.Permission) bool { return !t.Granted }), func(t *pb.Permission) string { return t.Domain })
	span.SetAttributes(attribute.String("key", req.EabKey), attribute.StringSlice("domains", req.Domains), attribute.String("user", key.User), attribute.StringSlice("missing", missing))
	api.logger.Info("Permissions checked", zap.Strings("missing", missing), zap.String("key", key.EabKey), zap.String("user", key.User))
	return &pb.CheckEABPermissionResponse{Missing: missing}, nil
}
