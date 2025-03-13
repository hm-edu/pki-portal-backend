package grpc

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	harica "github.com/hm-edu/harica/client"
	"github.com/hm-edu/harica/models"
	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/smimecertificate"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type smimeAPIServer struct {
	pb.UnimplementedSmimeServiceServer
	cfg              *cfg.PKIConfiguration
	logger           *zap.Logger
	validationClient *harica.Client
	db               *ent.Client
}

func newSmimeAPIServer(cfg *cfg.PKIConfiguration, db *ent.Client) (*smimeAPIServer, error) {
	validationClient, err := harica.NewClient(
		cfg.ValidationUser,
		cfg.ValidationPassword,
		cfg.ValidationTotpSeed,
		harica.WithRefreshInterval(5*time.Minute),
		harica.WithRetry(3),
	)
	if err != nil {
		return nil, err
	}
	return &smimeAPIServer{
			cfg:              cfg,
			logger:           zap.L(),
			validationClient: validationClient,
			db:               db,
		},
		nil
}

func (s *smimeAPIServer) ListCertificates(ctx context.Context, req *pb.ListSmimeRequest) (*pb.ListSmimeResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
		hub.Scope().SetUser(sentry.User{Email: req.Email})
	}
	logger := log.With(zap.String("user", req.Email))
	certs, err := s.db.SmimeCertificate.Query().Where(smimecertificate.EmailEQ(req.Email)).All(ctx)
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error fetching issued certs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching smime certificates")
	}
	resp := pb.ListSmimeResponse{}
	for _, cert := range certs {
		resp.Certificates = append(resp.Certificates, &pb.ListSmimeResponse_CertificateDetails{
			Serial:  cert.Serial,
			Expires: timestamppb.New(cert.NotAfter),
			Status:  string(cert.Status),
			Id:      int32(cert.ID),
		})
	}
	return &resp, nil
}
func (s *smimeAPIServer) IssueCertificate(ctx context.Context, req *pb.IssueSmimeRequest) (*pb.IssueSmimeResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
		hub.Scope().SetUser(sentry.User{Email: req.Email})
	}
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Issuing new smime certificate", Category: "info"}, nil)

	logger := log.With(zap.String("user", req.Email))
	logger.Info("Issuing new smime certificate")
	block, _ := pem.Decode([]byte(req.Csr))

	// Validate the passed CSR to comply the server-side requirements (e.g. key-strength, key-type, etc.)
	// The "real" user-data will be filled in by sectigo so we can sort of ignore any data provided by the user and simply pass the CSR to sectigo
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error while parsing CSR", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		hub.CaptureException(err)
		logger.Error("Error while validating CSR", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	if csr.PublicKeyAlgorithm != x509.RSA {
		return nil, status.Error(codes.InvalidArgument, "Only RSA keys are supported")
	}

	// Get the public key from the CSR
	pubKey, ok := csr.PublicKey.(*rsa.PublicKey)
	size := pubKey.Size() * 8
	if !ok || fmt.Sprintf("%d", size) != s.cfg.SmimeKeyLength {
		logger.Warn("Invalid key length", zap.String("key_length", fmt.Sprintf("%d", size)))
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}
	groups, err := s.validationClient.GetOrganizationsBulk()
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error fetching organizations", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching organizations")
	}

	cert, err := s.validationClient.RequestSmimeBulkCertificates(groups[0].OrganizationID, models.SmimeBulkRequest{
		Email:        req.Email,
		FriendlyName: req.CommonName,
		CertType:     "email_only",
		CSR:          string(req.Csr),
	})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error requesting certificate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error requesting certificate")
	}

	block, _ = pem.Decode([]byte(cert.Certificate))
	if block == nil {
		hub.CaptureException(err)
		logger.Error("Error decoding certificate")
		return nil, status.Error(codes.Internal, "Error decoding certificate")
	}
	certX509, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error parsing certificate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error parsing certificate")
	}

	s.db.SmimeCertificate.Create().
		SetCreateTime(time.Now()).
		SetEmail(req.Email).
		SetSerial(certX509.SerialNumber.String()).
		SetNotAfter(certX509.NotAfter).
		SetNotBefore(certX509.NotBefore).
		SetStatus(smimecertificate.StatusIssued).
		SetTransactionId(cert.TransactionID).Save(ctx)
	return &pb.IssueSmimeResponse{
		Certificate: cert.Certificate,
	}, nil
}
func (s *smimeAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSmimeRequest) (*emptypb.Empty, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	logger := log.With(zap.String("reason", req.Reason))
	logger.Info("Revoking smime certificate")

	return nil, status.Errorf(codes.Unimplemented, "method RevokeCertificate not implemented")
}
