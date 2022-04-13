package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
)

type domainAPIServer struct {
	pb.UnimplementedDomainServiceServer
	store *store.DomainStore
}

func newDomainAPIServer(store *store.DomainStore) *domainAPIServer {
	return &domainAPIServer{store: store}
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
	domains, err := api.store.ListAllDomains(ctx, true)
	if err != nil {
		return nil, err
	}
	missing := helper.Where(req.Domains, func(t string) bool {
		return !helper.Any(domains, func(d string) bool {
			return d == t
		})
	})
	return &pb.CheckRegistrationResponse{Missing: missing}, nil
}

func (api *domainAPIServer) ListDomains(ctx context.Context, req *pb.ListDomainsRequest) (*pb.ListDomainsResponse, error) {
	domains, err := api.store.ListDomains(ctx, req.User, req.Approved)
	if err != nil {
		return nil, err
	}

	resp := pb.ListDomainsResponse{Domains: helper.Map(domains, func(t *ent.Domain) string { return t.Fqdn })}

	return &resp, nil
}
