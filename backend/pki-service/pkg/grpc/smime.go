package grpc

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type smimeAPIServer struct {
	pb.UnimplementedSmimeServiceServer
	cfg    *cfg.PKIConfiguration
	logger *zap.Logger
}

func newSmimeAPIServer(cfg *cfg.PKIConfiguration) *smimeAPIServer {
	return &smimeAPIServer{cfg: cfg, logger: zap.L()}
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

	logger.Info("Requesting issued smime certificates")
	return nil, status.Errorf(codes.Unimplemented, "method ListCertificates not implemented")

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
	return nil, status.Errorf(codes.Unimplemented, "method IssueCertificate not implemented")

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
