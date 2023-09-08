package grpc

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/hm-edu/pki-service/pkg/cfg"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/client"
	"github.com/hm-edu/sectigo-client/sectigo/person"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	tracer trace.Tracer
}

func newSmimeAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration) *smimeAPIServer {
	tracer := otel.GetTracerProvider().Tracer("smime")
	return &smimeAPIServer{client: client, cfg: cfg, logger: zap.L(), tracer: tracer}
}

func (s *smimeAPIServer) ListCertificates(ctx context.Context, req *pb.ListSmimeRequest) (*pb.ListSmimeResponse, error) {
	_, span := s.tracer.Start(ctx, "listCertificates")
	defer span.End()
	logger := s.logger.With(zap.String("user", req.Email), zap.String("trace_id", span.SpanContext().TraceID().String()))

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
		span.RecordError(err)
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
	_, span := s.tracer.Start(ctx, "handleCsr")
	defer span.End()
	span.AddEvent("Validating csr")

	logger := s.logger.With(zap.String("user", req.Email), zap.String("trace_id", span.SpanContext().TraceID().String()))
	logger.Info("Issuing new smime certificate")
	block, _ := pem.Decode([]byte(req.Csr))

	// Validate the passed CSR to comply the server-side requirements (e.g. key-strength, key-type, etc.)
	// The "real" user-data will be filled in by sectigo so we can sort of ignore any data provided by the user and simply pass the CSR to sectigo
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		span.RecordError(err)
		logger.Error("Error while parsing CSR", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "Invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		span.RecordError(err)
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
		span.RecordError(err)
		logger.Error("Error while requesting person", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error requesting person")
	}
	if len(*persons) == 0 {
		logger.Info("No person found. Creating new")
		span.AddEvent("Creating new person")
		err = s.client.PersonService.CreatePerson(person.CreateRequest{
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			Email:          req.Email,
			OrganizationID: s.cfg.SmimeOrgID,
			ValidationType: "HIGH",
			CommonName:     req.CommonName,
			Phone:          "",
		})
		if err != nil {
			span.RecordError(err)
			logger.Error("Error while creating person", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error creating person")
		}
	} else {
		personItem := (*persons)[0]
		if personItem.ValidationType != "HIGH" {
			err := s.client.PersonService.UpdatePerson(personItem.ID, person.UpdateRequest{
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				OrganizationID: s.cfg.SmimeOrgID,
				ValidationType: "HIGH",
				CommonName:     req.CommonName,
			})
			if err != nil {
				span.RecordError(err)
				logger.Error("Error while updating person", zap.Error(err))
				return nil, status.Error(codes.Internal, "Error updating person")
			}
		}
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
		Term:            term,
		Eppn:            "",
	})
	if err != nil {
		span.RecordError(err)
		logger.Error("Error while enrolling certificate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error enrolling certificate")
	}
	cert := ""
	err = helper.WaitFor(10*time.Minute, 15*time.Second, func() (bool, error) {
		c, err := s.client.ClientService.Collect(resp.OrderNumber, "x509R")
		if err != nil {
			if e, ok := err.(*sectigo.ErrorResponse); ok {
				if e.Code == 0 && e.Description == "Being processed by Sectigo" {
					span.AddEvent("Certificate not ready yet")
					logger.Debug("Certificate not ready", zap.Int("id", resp.OrderNumber), zap.String("email", req.Email))
					return false, nil
				}
			}
			return false, err
		}
		span.AddEvent("Certificate ready")
		logger.Info("Certificate ready", zap.Int("id", resp.OrderNumber))
		cert = *c
		return true, nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.Internal, "Error obtaining certificate")
	}

	return &pb.IssueSmimeResponse{Certificate: cert}, nil

}
func (s *smimeAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSmimeRequest) (*emptypb.Empty, error) {
	_, span := s.tracer.Start(ctx, "listCertificates")
	defer span.End()

	logger := s.logger.With(zap.String("trace_id", span.SpanContext().TraceID().String()), zap.String("reason", req.Reason))

	switch req.Identifier.(type) {
	case *pb.RevokeSmimeRequest_Email:
		logger = logger.With(zap.String("email", req.GetEmail()))
		logger.Info("Revoking smime certificate")
		span.SetAttributes(attribute.String("email", req.GetEmail()))
		err := s.client.ClientService.RevokeByEmail(client.RevokeByEmailRequest{Email: req.GetEmail(), Reason: req.GetReason()})
		if err != nil {
			span.RecordError(err)
			logger.Error("Error while revoking certificate", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		logger.Info("Successfully revoked smime certificate")
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
