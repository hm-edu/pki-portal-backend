package grpc

import (
	"context"

	"github.com/hm-edu/dns-service/pkg/core"
	pb "github.com/hm-edu/portal-apis"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DNSServer struct {
	pb.UnimplementedDNSServiceServer
	logger   *zap.Logger
	provider core.DNSProvider
}

func NewDNSServer(logger *zap.Logger, provider core.DNSProvider) *DNSServer {
	return &DNSServer{
		logger:   logger,
		provider: provider,
	}
}

func (s *DNSServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	return nil, nil
}

func (s *DNSServer) Add(ctx context.Context, req *pb.AddRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *DNSServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *DNSServer) Update(ctx context.Context, req *pb.UpdateRequest) (*emptypb.Empty, error) {
	return nil, nil
}
