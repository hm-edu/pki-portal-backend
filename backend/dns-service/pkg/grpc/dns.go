package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hm-edu/dns-service/pkg/core"
	pb "github.com/hm-edu/portal-apis"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DNSServer is a DNS server.
type DNSServer struct {
	pb.UnimplementedDNSServiceServer
	logger   *zap.Logger
	provider core.DNSProvider
}

// NewDNSServer creates a new DNS server.
func NewDNSServer(logger *zap.Logger, provider core.DNSProvider) *DNSServer {
	return &DNSServer{
		logger:   logger,
		provider: provider,
	}
}

// List returns all the records for the given zone.
func (s *DNSServer) List(_ context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	logger := s.logger.With(zap.String("zone", req.Zone))
	logger.Info("listing zone")
	rrs, err := s.provider.List(req.Zone)
	if err != nil {
		logger.Error("failed to list RR", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list RR: %v", err))
	}
	entries := []*pb.DNSRecord{}
	for _, e := range rrs {
		x := s.parseRR(e)
		if x != nil {
			entries = append(entries, x)
		}
	}
	return &pb.ListResponse{Records: entries}, nil
}

func (s *DNSServer) parseRR(rr dns.RR) *pb.DNSRecord {
	if rr.Header().Class != dns.ClassINET {
		return nil
	}

	rrName := rr.Header().Name
	rrTTL := rr.Header().Ttl
	var rrType string
	var rrValues string
	switch rr.Header().Rrtype {
	case dns.TypeCNAME:
		rrValues = rr.(*dns.CNAME).Target
		rrType = "CNAME"
	case dns.TypeA:
		rrValues = rr.(*dns.A).A.String()
		rrType = "A"
	case dns.TypeAAAA:
		rrValues = rr.(*dns.AAAA).AAAA.String()
		rrType = "AAAA"
	case dns.TypeTXT:
		rrValues = strings.Join(rr.(*dns.TXT).Txt, "\n")
		rrType = "TXT"
	case dns.TypeNS:
		rrValues = rr.(*dns.NS).Ns
		rrType = "NS"
	case dns.TypeMX:
		rrValues = rr.(*dns.MX).Mx
		rrType = "MX"
	case dns.TypeNSEC, dns.TypeNSEC3, dns.TypeRRSIG, dns.TypeDNSKEY:
		s.logger.Debug("Ignoring DNSSEC record", zap.String("record", rr.String()))
		return nil
	case dns.TypeSOA:
		s.logger.Debug("Ignoring SOA record", zap.String("record", rr.String()))
		return nil
	default:
		s.logger.Warn("Ignoring record", zap.String("record", rr.String()))
		return nil
	}
	return &pb.DNSRecord{
		Name:    rrName,
		Ttl:     int32(rrTTL),
		Type:    rrType,
		Content: rrValues,
	}

}

func (s *DNSServer) buildRRs(rrs []*pb.DNSRecord, logger *zap.Logger) ([]dns.RR, error) {
	rr := []dns.RR{}

	for _, r := range rrs {
		newRR := fmt.Sprintf("%s %d %s %s", r.Name, r.Ttl, r.Type, r.Content)
		item, err := dns.NewRR(newRR)
		if err != nil {
			logger.Error("failed to build RR", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to build new RR: %v", err))
		}
		rr = append(rr, item)
	}
	return rr, nil
}

// Add adds the given records to the given zone.
func (s *DNSServer) Add(_ context.Context, req *pb.AddRequest) (*emptypb.Empty, error) {
	logger := s.logger.With(zap.String("zone", req.Zone), zap.Any("records", req.Records))
	logger.Info("adding records to zone")
	rrs, err := s.buildRRs(req.Records, logger)
	if err != nil {
		return nil, err
	}

	err = s.provider.Add(req.Zone, rrs)
	if err != nil {
		logger.Error("failed to add new RR", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add new RR: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// Delete deletes the given records from the given zone.
func (s *DNSServer) Delete(_ context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	logger := s.logger.With(zap.String("zone", req.Zone), zap.Any("records", req.Records))
	logger.Info("deleting records from zone")
	rrs, err := s.buildRRs(req.Records, logger)
	if err != nil {
		return nil, err
	}
	err = s.provider.Delete(req.Zone, rrs)
	if err != nil {
		logger.Error("failed to delete RR", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete RR: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// Update updates the given records in the given zone.
func (s *DNSServer) Update(_ context.Context, req *pb.UpdateRequest) (*emptypb.Empty, error) {
	for _, updateSet := range req.Updates {
		logger := s.logger.With(zap.String("zone", req.Zone), zap.Any("add", updateSet.New), zap.Any("delete", updateSet.Old))
		logger.Info("updating records in zone")
		oldRRs, err := s.buildRRs([]*pb.DNSRecord{updateSet.Old}, logger)
		if err != nil {
			return nil, err
		}
		newRRs, err := s.buildRRs([]*pb.DNSRecord{updateSet.New}, logger)
		if err != nil {
			return nil, err
		}
		err = s.provider.Update(req.Zone, []core.UpdateSet{{Old: oldRRs, New: newRRs}})
		if err != nil {
			logger.Error("failed to update RR", zap.Error(err))
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update RR: %v", err))
		}
	}
	return &emptypb.Empty{}, nil
}
