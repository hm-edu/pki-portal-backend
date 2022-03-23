package grpc

import (
	"context"

	pb "github.com/hm-edu/domain-api"
)

type domainAPIServer struct {
	pb.UnimplementedDomainServiceServer
}

func (api *domainAPIServer) CheckPermission(context.Context, *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	resp := pb.CheckPermissionResponse{Permissions: []*pb.Permission{}}
	return &resp, nil
}
