package grpc

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme"

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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
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
	client *sectigo.Client
	db     *ent.Client
	cfg    *cfg.SectigoConfiguration
	logger *zap.Logger

	acmeClient *acme.Client

	pendingValidations map[string]interface{}

	last     *time.Time
	duration *time.Duration
}

func newSslAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration, db *ent.Client) *sslAPIServer {
	keyFile := filepath.Join(cfg.AcmeStorage, "acme.key")

	var acmeClient *acme.Client
	var err error
	ctx := context.Background()

	sectigoDiectory := "https://acme.sectigo.com/v2/OV"

	acmeClient, err = pkiHelper.LoadAccount(ctx, keyFile, sectigoDiectory)
	if err != nil {
		zap.L().Info("Found no ACME account, creating one.")
		hmac, err := base64.RawURLEncoding.DecodeString(cfg.EabHmac)
		if err != nil {
			zap.L().Fatal("Failed to decode hmac", zap.Error(err))
		}

		acmeClient, err = pkiHelper.RegisterAccount(ctx, keyFile, sectigoDiectory, acme.ExternalAccountBinding{KID: cfg.EabKid, Key: hmac})
		if err != nil {
			zap.L().Fatal("Error registering ACME account", zap.Error(err))
		}
		zap.L().Info("Registered ACME account", zap.String("kid", cfg.EabKid))
	}

	gauge, _ := meter.AsyncInt64().Gauge(
		"ssl.issue.last.duration",
		instrument.WithUnit("seconds"),
		instrument.WithDescription("Required time for last SSL Certificates"),
	)

	gaugeLast, _ := meter.AsyncInt64().Gauge(
		"ssl.issue.last.unix",
		instrument.WithUnit("unixMilli"),
		instrument.WithDescription("Issue timestamp for last SSL Certificates"),
	)
	instance := &sslAPIServer{client: client, acmeClient: acmeClient, cfg: cfg, logger: zap.L(), db: db, pendingValidations: make(map[string]interface{})}
	err = meter.RegisterCallback([]instrument.Asynchronous{gauge, gaugeLast}, func(ctx context.Context) {
		if instance.last != nil {
			gauge.Observe(ctx, int64(instance.duration.Seconds()))
			gaugeLast.Observe(ctx, instance.last.UnixMilli())
		}
	})
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
		var created *timestamppb.Timestamp
		if x.Created != nil {
			created = timestamppb.New(*x.Created)
		}
		return &pb.SslCertificateDetails{
			Id:                      int32(x.SslId),
			CommonName:              x.CommonName,
			SubjectAlternativeNames: helper.Map(x.Edges.Domains, func(t *ent.Domain) string { return t.Fqdn }),
			Serial:                  x.Serial,
			Expires:                 timestamppb.New(x.NotAfter),
			NotBefore:               nbf,
			Status:                  string(x.Status),
			IssuedBy:                x.IssuedBy,
			Created:                 created,
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
	s.logger.Info("Issuing certificate",
		zap.String("common_name", csr.Subject.CommonName),
		zap.Strings("subject_alternative_names", sans))

	span.SetAttributes(attribute.StringSlice("subject_alternative_names", sans))
	for _, fqdn := range sans {
		id, err := s.db.Domain.Create().
			SetFqdn(fqdn).
			OnConflictColumns(domain.FieldFqdn).
			Ignore().
			ID(ctx)

		if err != nil {
			return s.handleError("Error while creating certificate", span, err)
		}
		ids = append(ids, id)
	}

	entry, err := s.db.Certificate.Create().
		SetCommonName(sans[0]).
		SetIssuedBy(req.Issuer).
		AddDomainIDs(ids...).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while creating certificate", span, err)
	}

	start := time.Now()
	certificates, err := pkiHelper.RequestCertificate(ctx, span, s.acmeClient, csr, sans)
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}

	stop := time.Now()
	duration := stop.Sub(start)
	s.duration = &duration
	s.last = &stop

	if err != nil {
		s.logger.Error("Error while registering callback", zap.Error(err))
	}
	s.logger.Info("Certificate issued", zap.Strings("sans", sans))
	certs, err := pkiHelper.LoadDER(certificates)
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}
	pem := certs[0]
	serial := fmt.Sprintf("%032x", pem.SerialNumber)
	s.logger.Info("Certificate issued",
		zap.Strings("subject_alternative_names", req.SubjectAlternativeNames),
		zap.Duration("duration", stop.Sub(start)),
		zap.String("certificate", serial))

	_, err = s.db.Certificate.UpdateOneID(entry.ID).
		SetSerial(pkiHelper.NormalizeSerial(serial)).
		SetStatus(certificate.StatusIssued).
		SetNotAfter(pem.NotAfter).
		SetNotBefore(pem.NotBefore).
		SetCreated(stop).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while saving collected certificate", span, err)
	}

	go func() {
		err := helper.WaitFor(10*time.Minute, 10*time.Second, func() (bool, error) {
			data, _, err := s.client.SslService.List(&ssl.ListSSLRequest{SerialNumber: serial})
			if err != nil {
				s.logger.Error("Error while listing certificates", zap.Error(err))
				return false, err
			}
			if len(*data) == 0 {
				s.logger.Debug("No certificates found")
				return false, nil
			}
			cert := (*data)[0]
			_, err = s.db.Certificate.UpdateOneID(entry.ID).SetSslId(cert.SslID).Save(context.Background())
			if err != nil {
				s.logger.Error("Error while updating certificate", zap.Error(err))
				return true, err
			}
			return true, nil
		})
		if err != nil {
			s.logger.Error("Error while extending information for certificate",
				zap.String("certificate", serial),
				zap.Error(err))
		}

	}()

	return &pb.IssueSslResponse{Certificate: flattenCertificates(certs)}, nil
}

func flattenCertificates(certs []*x509.Certificate) string {
	var result []byte
	for _, cert := range certs {
		c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		result = append(result, c...)
	}
	return string(result)
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
		certs, err := s.db.Certificate.Query().
			Where(certificate.And(certificate.HasDomainsWith(domain.FqdnEQ(req.GetCommonName())),
				certificate.StatusNEQ(certificate.StatusRevoked),
				certificate.NotAfterGT(time.Now()))).
			All(ctx)

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
