package grpc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/ent/predicate"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pkiHelper "github.com/hm-edu/pki-service/pkg/helper"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/ssl"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	tracingCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var meter = otel.GetMeterProvider().Meter("pki-service")

func mapCertificate(x *ent.Certificate) *pb.SslCertificateDetails {

	var nbf *timestamppb.Timestamp
	if x.NotBefore != nil {
		nbf = timestamppb.New(*x.NotBefore)
	}
	var created *timestamppb.Timestamp
	if x.Created != nil {
		created = timestamppb.New(*x.Created)
	}
	issuedBy := ""
	if x.IssuedBy != nil {
		issuedBy = *x.IssuedBy
	}
	source := ""
	if x.Source != nil {
		source = *x.Source
	}
	return &pb.SslCertificateDetails{
		Id:                      int32(x.SslId),
		DbId:                    int32(x.ID),
		CommonName:              x.CommonName,
		SubjectAlternativeNames: helper.Map(x.Edges.Domains, func(t *ent.Domain) string { return t.Fqdn }),
		Serial:                  x.Serial,
		Expires:                 timestamppb.New(x.NotAfter),
		NotBefore:               nbf,
		Status:                  string(x.Status),
		Source:                  source,
		IssuedBy:                issuedBy,
		Created:                 created,
	}
}

func flattenCertificates(certs []*x509.Certificate) string {
	var result []byte
	for _, cert := range certs {
		c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		result = append(result, c...)
	}
	return string(result)
}

func (s *sslAPIServer) handleError(msg string, span trace.Span, err error, logger *zap.Logger) (*pb.IssueSslResponse, error) {
	span.RecordError(err)
	span.AddEvent(msg)
	logger.Error(msg, zap.Error(err))
	return nil, status.Errorf(codes.Internal, msg)
}

type sslAPIServer struct {
	pb.UnimplementedSSLServiceServer
	client *sectigo.Client
	db     *ent.Client
	cfg    *cfg.SectigoConfiguration
	logger *zap.Logger

	pendingValidations map[string]interface{}

	last     *time.Time
	duration *time.Duration
}

func newSslAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration, db *ent.Client) *sslAPIServer {
	var err error

	gauge, _ := meter.Int64ObservableGauge(
		"ssl.issue.last.duration",
		metric.WithUnit("seconds"),
		metric.WithDescription("Required time for last SSL Certificates"),
	)

	gaugeLast, _ := meter.Int64ObservableGauge(
		"ssl.issue.last.unix",
		metric.WithUnit("unixMilli"),
		metric.WithDescription("Issue timestamp for last SSL Certificates"),
	)
	instance := &sslAPIServer{client: client, cfg: cfg, logger: zap.L(), db: db, pendingValidations: make(map[string]interface{})}
	_, err = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		if instance.last != nil {
			observer.ObserveInt64(gauge, int64(instance.duration.Seconds()))
			observer.ObserveInt64(gaugeLast, instance.last.UnixMilli())
		}
		return nil
	}, gauge, gaugeLast)
	if err != nil {
		zap.L().Error("Failed to register callback", zap.Error(err))
	}
	return instance
}

func (s *sslAPIServer) CertificateDetails(ctx context.Context, req *pb.CertificateDetailsRequest) (*pb.SslCertificateDetails, error) {
	x, err := s.db.Certificate.Query().WithDomains().Where(certificate.Serial(req.Serial)).First(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Certificate not found")
	}
	return mapCertificate(x), nil
}

func (s *sslAPIServer) ListCertificates(ctx context.Context, req *pb.ListSslRequest) (*pb.ListSslResponse, error) {

	_, span := otel.GetTracerProvider().Tracer("ssl").Start(ctx, "listing ssl certificates")
	defer span.End()
	logger := s.logger.With(zap.Strings("domains", req.Domains), zap.Bool("partial", req.IncludePartial))
	logger.Debug("listing certificates for domains")

	var cond predicate.Certificate
	if req.IncludePartial {
		cond = certificate.HasDomainsWith(domain.FqdnIn(req.Domains...))
	} else {
		cond = certificate.And(
			certificate.HasDomainsWith(domain.FqdnIn(req.Domains...)),
			certificate.Not(certificate.HasDomainsWith(domain.FqdnNotIn(req.Domains...))),
		)
	}
	certificates, err := s.db.Certificate.Query().WithDomains().Where(cond).All(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(tracingCodes.Error, "listing certificates failed")
		return nil, status.Error(codes.Internal, "Error querying certificates")
	}
	return &pb.ListSslResponse{Items: helper.Map(certificates, mapCertificate)}, nil
}

func (s *sslAPIServer) IssueCertificate(ctx context.Context, req *pb.IssueSslRequest) (*pb.IssueSslResponse, error) {

	_, span := otel.GetTracerProvider().Tracer("ssl").Start(ctx, "issuing server certificate")
	defer span.End()
	logger := s.logger.With(zap.String("trace_id", span.SpanContext().TraceID().String()), zap.String("issuer", req.Issuer))

	block, _ := pem.Decode([]byte(req.Csr))

	if block == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid pem block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		span.RecordError(err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		span.RecordError(err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid CSR signature")
	}
	var sans []string
	if csr.Subject.CommonName != "" {
		sans = []string{csr.Subject.CommonName}
	}

	for _, domain := range req.SubjectAlternativeNames {
		if domain != csr.Subject.CommonName && domain != "" {
			sans = append(sans, domain)
		}
	}
	ids := []int{}

	if _, ok := s.pendingValidations[req.Csr]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "Outstanding validation for this CSR")
	}
	logger = logger.With(zap.Strings("subject_alternative_names", sans))
	logger.Info("Issuing new server certificate")

	span.SetAttributes(attribute.StringSlice("subject_alternative_names", sans))
	for _, fqdn := range sans {
		id, err := s.db.Domain.Create().
			SetFqdn(fqdn).
			OnConflictColumns(domain.FieldFqdn).
			Ignore().
			ID(ctx)

		if err != nil {
			return s.handleError("Error while creating certificate", span, err, logger)
		}
		ids = append(ids, id)
	}

	entry, err := s.db.Certificate.Create().
		SetCommonName(sans[0]).
		SetIssuedBy(req.Issuer).
		SetSource(req.Source).
		AddDomainIDs(ids...).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while creating certificate", span, err, logger)
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
		return s.handleError("Error while requesting certificate", span, err, logger)
	}

	entry, err = s.db.Certificate.UpdateOneID(entry.ID).SetStatus(certificate.StatusRequested).SetSslId(enrollment.SslID).Save(ctx)
	if err != nil {
		return s.handleError("Error while storing certificate", span, err, logger)
	}
	cert := ""
	err = helper.WaitFor(5*time.Minute, 1*time.Second, func() (bool, error) {
		c, err := s.client.SslService.Collect(enrollment.SslID, "x509R")
		if err != nil {
			if e, ok := err.(*sectigo.ErrorResponse); ok {
				if e.Code == 0 && e.Description == "Being processed by Sectigo" {
					span.AddEvent("Certificate not ready yet")
					s.logger.Debug("Certificate not ready", zap.Int("id", enrollment.SslID), zap.Strings("subject_alternative_names", req.SubjectAlternativeNames))
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
		return s.handleError("Error collecting certificate", span, err, logger)
	}

	stop := time.Now()
	duration := stop.Sub(start)
	s.duration = &duration
	s.last = &stop

	certs, err := pkiHelper.ParseCertificates([]byte(cert))
	if err != nil {
		return s.handleError("Error parsing certificate", span, err, logger)
	}
	pem := certs[0]
	serial := fmt.Sprintf("%032x", pem.SerialNumber)
	logger.Info("Certificate issued",
		zap.Duration("duration", stop.Sub(start)),
		zap.String("serial", serial))

	_, err = s.db.Certificate.UpdateOneID(entry.ID).
		SetSerial(pkiHelper.NormalizeSerial(serial)).
		SetStatus(certificate.StatusIssued).
		SetNotAfter(pem.NotAfter).
		SetNotBefore(pem.NotBefore).
		SetCreated(stop).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while saving collected certificate", span, err, logger)
	}

	return &pb.IssueSslResponse{Certificate: flattenCertificates(certs)}, nil
}

func (s *sslAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSslRequest) (*emptypb.Empty, error) {

	_, span := otel.GetTracerProvider().Tracer("ssl").Start(ctx, "revoke ssl certificate")
	defer span.End()
	logger := s.logger.With(zap.String("trace_id", span.SpanContext().TraceID().String()), zap.String("reason", req.Reason))

	errorReturn := func(err error, logger *zap.Logger) (*emptypb.Empty, error) {
		logger.Error("Failed to revoke certificate", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to revoke certificate")
	}

	switch req.Identifier.(type) {
	case *pb.RevokeSslRequest_Serial:
		serial := req.GetSerial()
		logger := logger.With(zap.String("serial", serial))
		logger.Info("Revoking certificate by serial")
		c, err := s.db.Certificate.Query().Where(certificate.Serial(serial)).First(ctx)
		if err != nil {
			return errorReturn(err, logger)
		}
		err = s.client.SslService.RevokeBySslID(fmt.Sprint(c.SslId), req.Reason)
		if sectigoError, ok := err.(*sectigo.ErrorResponse); ok && sectigoError.Code == -102 {
			logger.Info("Certificate already revoked")
		} else if err != nil {
			logger.Error("Revoking request failed", zap.Error(err))
			return errorReturn(err, logger)
		}
		_, err = s.db.Certificate.UpdateOneID(c.ID).SetStatus(certificate.StatusRevoked).Save(ctx)
		if err != nil {
			return errorReturn(err, logger)
		}
	case *pb.RevokeSslRequest_CommonName:
		logger := logger.With(zap.String("common_name", req.GetCommonName()))
		logger.Info("Revoking certificate by common name")
		certs, err := s.db.Certificate.Query().
			Where(certificate.And(certificate.HasDomainsWith(domain.FqdnEQ(req.GetCommonName())),
				certificate.StatusNEQ(certificate.StatusRevoked),
				certificate.StatusNEQ(certificate.StatusInvalid),
				certificate.NotAfterGT(time.Now()))).
			All(ctx)

		if err != nil {
			return nil, status.Error(codes.Internal, "Error querying certificates")
		}

		ret := make(chan struct{ err error }, len(certs))

		for _, c := range certs {
			go func(c *ent.Certificate, ret chan struct{ err error }) {
				err := s.client.SslService.RevokeBySslID(fmt.Sprint(c.SslId), req.Reason)
				if err != nil {
					if sectigoError, ok := err.(*sectigo.ErrorResponse); ok && sectigoError.Code == -102 {
						logger.Info("Certificate already revoked")
					} else {
						logger.Error("Revoking request failed", zap.Error(err))
						ret <- struct{ err error }{err}
						return
					}
				}
				_, err = s.db.Certificate.UpdateOneID(c.ID).SetStatus(certificate.StatusRevoked).Save(ctx)
				if err != nil {
					ret <- struct{ err error }{err}
					return
				}
				ret <- struct{ err error }{}
			}(c, ret)
		}
		var errors []error
		for i := 0; i < len(certs); i++ {
			select {
			case err := <-ret:
				if err.err != nil {
					errors = append(errors, err.err)
				}
			case <-ctx.Done():
				return nil, status.Error(codes.Canceled, "Canceled")
			}
		}
		if len(errors) > 0 {
			return errorReturn(fmt.Errorf("%v", errors), logger)
		}
	}
	return &emptypb.Empty{}, nil
}
