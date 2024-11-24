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
	"github.com/hm-edu/pki-service/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/client"
	"github.com/hm-edu/sectigo-client/sectigo/person"

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

func (s *smimeAPIServer) ListCertificates(ctx context.Context, req *pb.ListSmimeRequest) (*pb.ListSmimeResponse, error) {
	span := sentry.StartSpan(ctx, "List S/MIME Certificates")
	defer span.Finish()
	ctx = span.Context()
	hub := sentry.GetHubFromContext(ctx)
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}
	logger := log.With(zap.String("user", req.Email))

	logger.Info("Requesting issued smime certificates")
	items, err := s.client.ClientService.ListByEmail(req.Email)
	if err != nil {
		if sectigoError, ok := err.(*sectigo.ErrorResponse); ok {
			if sectigoError.Code == -105 {
				logger.Info("No certificates found")
				return &pb.ListSmimeResponse{Certificates: []*pb.ListSmimeResponse_CertificateDetails{}}, nil
			}
		}
		logger.Info("Error while requesting smime certificates", zap.Error(err))
		hub.CaptureException(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.Info("Successfully requested smime certificates", zap.Int("count", len(*items)))
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
	span := sentry.StartSpan(ctx, "Issue S/MIME Certificate")
	defer span.Finish()
	ctx = span.Context()
	hub := sentry.GetHubFromContext(ctx)
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
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

	term := s.cfg.SmimeTerm
	if req.Student {
		term = s.cfg.SmimeStudentTerm
	}

	persons, err := s.client.PersonService.List(&person.ListParams{Email: req.Email})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error while requesting person", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error requesting person")
	}

	validationLevel := "HIGH"
	profile := s.cfg.SmimeProfile

	if req.ValidationStandard {
		validationLevel = "STANDARD"
		profile = s.cfg.SmimeProfileStandard
		if profile == 0 {
			logger.Warn("No profile for validation level standard configured")
			return nil, status.Error(codes.InvalidArgument, "Validation level standard not supported")
		}
	}

	if len(*persons) == 0 {
		logger.Info("No person found. Creating new")
		err = s.client.PersonService.CreatePerson(person.CreateRequest{
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			Email:          req.Email,
			OrganizationID: s.cfg.SmimeOrgID,
			ValidationType: validationLevel,
			CommonName:     req.CommonName,
			Phone:          "",
		})
		if err != nil {
			logger.Error("Error while creating person", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error creating person")
		}
	} else {
		personItem := (*persons)[0]
		if personItem.ValidationType != "HIGH" && !req.ValidationStandard {
			err := s.client.PersonService.UpdatePerson(personItem.ID, person.UpdateRequest{
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				OrganizationID: s.cfg.SmimeOrgID,
				ValidationType: validationLevel,
				CommonName:     req.CommonName,
			})
			if err != nil {
				hub.CaptureException(err)
				logger.Error("Error while updating person", zap.Error(err))
				return nil, status.Error(codes.Internal, "Error updating person")
			}
		}
	}

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
		CertType:        profile,
		Term:            term,
		Eppn:            "",
	})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error while enrolling certificate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error enrolling certificate")
	}
	cert := ""
	err = helper.WaitFor(10*time.Minute, 15*time.Second, func() (bool, error) {
		c, err := s.client.ClientService.Collect(resp.OrderNumber, "x509R")
		if err != nil {
			if e, ok := err.(*sectigo.ErrorResponse); ok {
				if e.Code == 0 && e.Description == "Being processed by Sectigo" {
					logger.Debug("Certificate not ready", zap.Int("id", resp.OrderNumber), zap.String("email", req.Email))
					return false, nil
				}
			}
			return false, err
		}
		logger.Info("Certificate ready", zap.Int("id", resp.OrderNumber))
		cert = *c
		return true, nil
	})

	if err != nil {
		hub.CaptureException(err)
		return nil, status.Error(codes.Internal, "Error obtaining certificate")
	}

	return &pb.IssueSmimeResponse{Certificate: cert}, nil

}
func (s *smimeAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSmimeRequest) (*emptypb.Empty, error) {
	span := sentry.StartSpan(ctx, "Revoke S/MIME Certificate")
	defer span.Finish()
	ctx = span.Context()
	hub := sentry.GetHubFromContext(ctx)
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	logger := log.With(zap.String("reason", req.Reason))

	switch req.Identifier.(type) {
	case *pb.RevokeSmimeRequest_Email:
		logger = logger.With(zap.String("email", req.GetEmail()))
		logger.Info("Revoking smime certificate")
		err := s.client.ClientService.RevokeByEmail(client.RevokeByEmailRequest{Email: req.GetEmail(), Reason: req.GetReason()})
		if err != nil {
			hub.CaptureException(err)
			logger.Error("Error while revoking certificate", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		logger.Info("Successfully revoked smime certificate")
		return &emptypb.Empty{}, nil
	case *pb.RevokeSmimeRequest_Serial:
		err := s.client.ClientService.RevokeBySerial(client.RevokeBySerialRequest{Serial: req.GetSerial(), Reason: req.GetReason()})
		if err != nil {
			hub.CaptureException(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &emptypb.Empty{}, nil
	}
	return nil, status.Errorf(codes.Unimplemented, "method RevokeCertificate not implemented")
}
