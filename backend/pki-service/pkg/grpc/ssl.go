package grpc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pkiHelper "github.com/hm-edu/pki-service/pkg/helper"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/ssl"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var meter = global.MeterProvider().Meter("pki-service")

type sslAPIServer struct {
	pb.UnimplementedSSLServiceServer
	client   *sectigo.Client
	db       *ent.Client
	cfg      *cfg.SectigoConfiguration
	logger   *zap.Logger
	duration syncint64.Histogram
}

func newSslAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration, db *ent.Client) *sslAPIServer {

	durRecorder, _ := meter.SyncInt64().Histogram(
		"ssl.issue.duration",
		instrument.WithUnit("milliseconds"),
		instrument.WithDescription("Issue time for SSL Certificates"),
	)

	return &sslAPIServer{client: client, cfg: cfg, logger: zap.L(), db: db, duration: durRecorder}
}

func parseCertificates(cert []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for block, rest := pem.Decode(cert); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			certs = append(certs, cert)
		default:
			return nil, errors.New("Unknown entry in cert chain")
		}
	}
	return certs, nil
}

func (s *sslAPIServer) CertificateDetails(ctx context.Context, req *pb.CertificateDetailsRequest) (*pb.SslCertificateDetails, error) {
	x, err := s.db.Certificate.Query().WithDomains().Where(certificate.Serial(req.Serial)).First(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Certificate not found")
	}
	var nbf *timestamppb.Timestamp
	if x.NotBefore != nil {
		nbf = timestamppb.New(*x.NotBefore)
	}
	return &pb.SslCertificateDetails{
		Id:                      int32(x.SslId),
		CommonName:              x.CommonName,
		SubjectAlternativeNames: helper.Map(x.Edges.Domains, func(t *ent.Domain) string { return t.Fqdn }),
		Serial:                  x.Serial,
		Expires:                 timestamppb.New(x.NotAfter),
		NotBefore:               nbf,
	}, nil
}

func (s *sslAPIServer) ListCertificates(ctx context.Context, req *pb.ListSslRequest) (*pb.ListSslResponse, error) {
	certificates, err := s.db.Certificate.Query().WithDomains().Where(certificate.HasDomainsWith(domain.FqdnIn(req.Domains...))).All(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error querying certificates")
	}
	return &pb.ListSslResponse{Items: helper.Map(certificates, func(x *ent.Certificate) *pb.SslCertificateDetails {

		var nbf *timestamppb.Timestamp
		if x.NotBefore != nil {
			nbf = timestamppb.New(*x.NotBefore)
		}
		return &pb.SslCertificateDetails{
			Id:                      int32(x.SslId),
			CommonName:              x.CommonName,
			SubjectAlternativeNames: helper.Map(x.Edges.Domains, func(t *ent.Domain) string { return t.Fqdn }),
			Serial:                  x.Serial,
			Expires:                 timestamppb.New(x.NotAfter),
			NotBefore:               nbf,
		}
	})}, nil
}

func (s *sslAPIServer) IssueCertificate(ctx context.Context, req *pb.IssueSslRequest) (*pb.IssueSslResponse, error) {

	_, span := otel.GetTracerProvider().Tracer("ssl").Start(ctx, "handleCsr")
	defer span.End()

	block, _ := pem.Decode([]byte(req.Csr))

	if block == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSR")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		span.RecordError(err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		span.RecordError(err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSR")
	}
	sans := []string{csr.Subject.CommonName}

	for _, domain := range req.SubjectAlternativeNames {
		if domain != csr.Subject.CommonName {
			sans = append(sans, domain)
		}
	}
	ids := []int{}

	for _, fqdn := range sans {
		id, err := s.db.Domain.Create().SetFqdn(fqdn).OnConflictColumns(domain.FieldFqdn).Ignore().ID(ctx)

		if err != nil {
			return s.handleError("Error while creating certificate", span, err)
		}
		ids = append(ids, id)
	}
	entry, err := s.db.Certificate.Create().SetCommonName(csr.Subject.CommonName).AddDomainIDs(ids...).Save(ctx)
	if err != nil {
		return s.handleError("Error while creating certificate", span, err)
	}

	start := time.Now()

	enrollment, err := s.client.SslService.Enroll(ssl.EnrollmentRequest{
		OrgID:        s.cfg.SslOrgID,
		Csr:          req.Csr,
		Term:         s.cfg.SslTerm,
		CertType:     s.cfg.SslProfile,
		SubjAltNames: strings.Join(req.SubjectAlternativeNames, ","),
	})

	span.AddEvent("Enrollment request sent")
	if err != nil {
		return s.handleError("Error while requesting certificate", span, err)
	}

	entry, err = s.db.Certificate.UpdateOneID(entry.ID).SetStatus(certificate.StatusRequested).SetSslId(enrollment.SslID).Save(ctx)
	if err != nil {
		return s.handleError("Error while storing certificate", span, err)
	}

	cert := ""
	err = helper.WaitFor(10*time.Minute, 10*time.Second, func() (bool, error) {
		c, err := s.client.SslService.Collect(enrollment.SslID, "x509")
		if err != nil {
			if e, ok := err.(*sectigo.ErrorResponse); ok {
				if e.Code == 0 && e.Description == "Being processed by Sectigo" {
					span.AddEvent("Certificate not ready yet")
					s.logger.Info("Certificate not ready", zap.Int("id", enrollment.SslID), zap.Strings("subject_alternative_names", req.SubjectAlternativeNames))
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
		return s.handleError("Error while collecting certificate", span, err)
	}

	stop := time.Now()
	s.duration.Record(ctx, stop.Sub(start).Milliseconds())

	certs, err := parseCertificates([]byte(cert))
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}
	pem := certs[len(certs)-1]
	_, err = s.db.Certificate.UpdateOneID(entry.ID).SetSerial(pkiHelper.NormalizeSerial(pem.Subject.SerialNumber)).SetStatus(certificate.StatusIssued).SetNotAfter(pem.NotAfter).SetNotBefore(pem.NotBefore).Save(ctx)
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}

	return &pb.IssueSslResponse{Certificate: cert}, nil
}

func (s *sslAPIServer) handleError(msg string, span trace.Span, err error) (*pb.IssueSslResponse, error) {
	span.RecordError(err)
	span.AddEvent(msg)
	s.logger.Error(msg, zap.Error(err))
	return nil, status.Errorf(codes.Internal, msg)
}

func (s *sslAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSslRequest) (*emptypb.Empty, error) {

	errorReturn := func(err error) (*emptypb.Empty, error) {
		s.logger.Error("Failed to revoke certificate", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to revoke certificate")
	}

	switch req.Identifier.(type) {
	case *pb.RevokeSslRequest_Serial:
		serial := req.GetSerial()
		s.logger.Info("Revoking certificate by serial", zap.String("serial", serial))
		err := s.client.SslService.Revoke(serial, req.Reason)
		if err != nil {
			return errorReturn(err)
		}
	case *pb.RevokeSslRequest_CommonName:
		s.logger.Info("Revoking certificate by common name", zap.String("common_name", req.GetCommonName()))
		certs, err := s.db.Certificate.Query().Where(certificate.And(certificate.HasDomainsWith(domain.FqdnEQ(req.GetCommonName())), certificate.StatusNEQ(certificate.StatusRevoked), certificate.NotAfterGT(time.Now()))).All(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error querying certificates")
		}
		for _, c := range certs {
			err := s.client.SslService.Revoke(c.Serial, req.Reason)
			if err != nil {
				return errorReturn(err)
			}
			_, err = s.db.Certificate.UpdateOneID(c.ID).SetStatus(certificate.StatusRevoked).Save(ctx)
			if err != nil {
				return errorReturn(err)
			}
		}
	}
	return &emptypb.Empty{}, nil
}
