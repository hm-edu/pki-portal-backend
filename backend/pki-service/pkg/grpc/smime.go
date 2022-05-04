package grpc

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/hm-edu/pki-service/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/client"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type smimeAPIServer struct {
	pb.UnimplementedSmimeServiceServer
	client *sectigo.Client
	cfg    *cfg.SectigoConfiguration
	logger *zap.Logger
}

func newSmimeAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration) *smimeAPIServer {
	return &smimeAPIServer{client: client, cfg: cfg, logger: zap.L()}
}

func (s *smimeAPIServer) ListCertificates(_ context.Context, req *pb.ListSmimeRequest) (*pb.ListSmimeResponse, error) {
	s.logger.Debug("Requesting smime certificates", zap.String("user", req.Email))
	items, err := s.client.ClientService.ListByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ListSmimeResponse{Certificates: helper.Map(*items, func(t client.ListItem) *pb.ListSmimeResponse_CertificateDetails {
		return &pb.ListSmimeResponse_CertificateDetails{
			Id:      int32(t.ID),
			Serial:  t.SerialNumber,
			Status:  string(t.State),
			Expires: timestamppb.New(t.Expires.Time),
		}
	})}, nil
}
func (s *smimeAPIServer) IssueCertificate(ctx context.Context, req *pb.IssueSmimeRequest) (*pb.IssueSmimeResponse, error) {

	_, span := otel.GetTracerProvider().Tracer("smime").Start(ctx, "handleCsr")
	defer span.End()
	span.AddEvent("Validating csr")

	block, _ := pem.Decode([]byte(req.Csr))

	// Validate the passed CSR to comply the server-side requirements (e.g. key-strength, key-type, etc.)
	// The "real" user-data will be filled in by sectigo so we can sort of ignore any data provided by the user and simply pass the CSR to sectigo
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	if csr.PublicKeyAlgorithm != x509.RSA {
		return nil, status.Error(codes.InvalidArgument, "Only RSA keys are supported")

	}

	// Get the public key from the CSR
	pubKey, ok := csr.PublicKey.(*rsa.PublicKey)
	size := pubKey.Size() * 8
	if !ok || fmt.Sprintf("%d", size) != s.cfg.SmimeKeyLength {
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	span.AddEvent("Enrolling certificate")
	resp, err := s.client.ClientService.Enroll(client.EnrollmentRequest{
		OrgID:           s.cfg.SmimeOrgID,
		FirstName:       req.FirstName,
		MiddleName:      req.MiddleName,
		CommonName:      req.CommonName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           "",
		SecondaryEmails: []string{},
		CSR:             req.Csr,
		CertType:        s.cfg.SmimeProfile,
		Term:            s.cfg.SmimeTerm,
		Eppn:            "",
	})
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.Internal, "Error enrolling certificate")
	}
	cert := ""
	err = helper.WaitFor(5*time.Minute, 5*time.Second, func() (bool, error) {
		c, err := s.client.ClientService.Collect(resp.OrderNumber, "x509")
		if err != nil {
			if e, ok := err.(*sectigo.ErrorResponse); ok {
				if e.Code == 0 && e.Description == "Being processed by Sectigo" {
					span.AddEvent("Certificate not ready yet")
					s.logger.Info("Certificate not ready", zap.Int("id", resp.OrderNumber), zap.String("email", req.Email))
					return false, nil
				}
			}
			return false, err
		}
		span.AddEvent("Certificate ready")
		cert = *c
		return true, nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.Internal, "Error obtaining certificate")
	}

	cert = strings.ReplaceAll(cert, "-----BEGIN PKCS7-----\n", "")
	cert = strings.ReplaceAll(cert, "-----END PKCS7-----\n", "")

	return &pb.IssueSmimeResponse{Certificate: cert}, nil

}
func (s *smimeAPIServer) RevokeCertificate(_ context.Context, req *pb.RevokeSmimeRequest) (*emptypb.Empty, error) {

	switch req.Identifier.(type) {
	case *pb.RevokeSmimeRequest_Email:
		err := s.client.ClientService.RevokeByEmail(client.RevokeByEmailRequest{Email: req.GetEmail(), Reason: req.GetReason()})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &emptypb.Empty{}, nil
	case *pb.RevokeSmimeRequest_Serial:
		err := s.client.ClientService.RevokeBySerial(client.RevokeBySerialRequest{Serial: req.GetSerial(), Reason: req.GetReason()})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &emptypb.Empty{}, nil
	}
	return nil, status.Errorf(codes.Unimplemented, "method RevokeCertificate not implemented")
}
