package grpc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	legoCert "github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	legoLog "github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pkiHelper "github.com/hm-edu/pki-service/pkg/helper"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"

	"go.opentelemetry.io/otel"
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
	client  *sectigo.Client
	db      *ent.Client
	legoCfg *lego.Config
	cfg     *cfg.SectigoConfiguration
	logger  *zap.Logger

	pendingValidations map[string]interface{}

	last     time.Time
	duration time.Duration
}

func registerAcme(client *lego.Client, config *cfg.SectigoConfiguration, account pkiHelper.User, accountFile string, keyFile string) error {
	reg, err := client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
		TermsOfServiceAgreed: true,
		Kid:                  config.EabKid,
		HmacEncoded:          config.EabHmac,
	})
	if err != nil {
		return err
	}
	account.Registration = reg
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}
	err = os.WriteFile(accountFile, data, 0600)
	if err != nil {
		return err
	}
	certOut, err := os.OpenFile(keyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec
	if err != nil {
		return err
	}
	defer func(certOut *os.File) {
		_ = certOut.Close()
	}(certOut)

	pemKey := certcrypto.PEMBlock(account.Key)
	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return err
	}
	return nil
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(file) //#nosec
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private key type")
}
func newSslAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration, db *ent.Client) *sslAPIServer {
	accountFile := filepath.Join(cfg.AcmeStorage, "reg.json")
	keyFile := filepath.Join(cfg.AcmeStorage, "reg.key")

	var account pkiHelper.User
	if ok, _ := fileExists(accountFile); !ok {
		// Actually we would not need a private key but the lego API requires one.
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil
		}

		account = pkiHelper.User{
			Key: privateKey,
		}

	} else {
		data, err := os.ReadFile(accountFile) //#nosec
		if err != nil {
			return nil
		}
		err = json.Unmarshal(data, &account)
		if err != nil {
			return nil
		}
		account.Key, err = loadPrivateKey(keyFile)
		if err != nil {
			return nil
		}

	}
	legoCfg := lego.NewConfig(&account)
	legoCfg.CADirURL = "https://acme.sectigo.com/v2/OV"
	legoLog.Logger = pkiHelper.NewZapLogger(zap.L())
	legoCfg.Certificate.Timeout = time.Duration(5) * time.Minute
	if account.Registration == nil {
		legoClient, err := lego.NewClient(legoCfg)

		if err != nil {
			return nil
		}
		err = registerAcme(legoClient, cfg, account, accountFile, keyFile)
		if err != nil {
			return nil
		}
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
	instance := &sslAPIServer{client: client, legoCfg: legoCfg, cfg: cfg, logger: zap.L(), db: db, pendingValidations: make(map[string]interface{})}
	err := meter.RegisterCallback([]instrument.Asynchronous{s.gauge, s.gaugeLast}, func(ctx context.Context) {
		gauge.Observe(ctx, int64(instance.duration.Seconds()))
		gaugeLast.Observe(ctx, instance.last.UnixMilli())
	})
	if err != nil {
		zap.L().Error("Failed to register callback", zap.Error(err))
	}
	return instance
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
	s.logger.Debug("Issuing certificate", zap.String("common_name", csr.Subject.CommonName), zap.Strings("subject_alternative_names", sans))

	for _, fqdn := range sans {
		id, err := s.db.Domain.Create().SetFqdn(fqdn).OnConflictColumns(domain.FieldFqdn).Ignore().ID(ctx)

		if err != nil {
			return s.handleError("Error while creating certificate", span, err)
		}
		ids = append(ids, id)
	}
	entry, err := s.db.Certificate.Create().SetCommonName(sans[0]).AddDomainIDs(ids...).Save(ctx)
	if err != nil {
		return s.handleError("Error while creating certificate", span, err)
	}

	start := time.Now()
	client, err := lego.NewClient(s.legoCfg)
	if err != nil {
		return s.handleError("Error while creating client", span, err)
	}

	certificates, err := client.Certificate.ObtainForCSR(legoCert.ObtainForCSRRequest{CSR: csr, Bundle: true})
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}

	stop := time.Now()

	s.duration = stop.Sub(start)
	s.last = stop

	if err != nil {
		s.logger.Error("Error while registering callback", zap.Error(err))
	}
	certs, err := parseCertificates(certificates.Certificate)
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}
	pem := certs[0]
	s.logger.Info("Certificate issued", zap.Strings("subject_alternative_names", req.SubjectAlternativeNames), zap.Duration("duration", stop.Sub(start)), zap.String("certificate", fmt.Sprintf("%032x", pem.SerialNumber)))
	_, err = s.db.Certificate.UpdateOneID(entry.ID).SetSerial(pkiHelper.NormalizeSerial(fmt.Sprintf("%032x", pem.SerialNumber))).SetStatus(certificate.StatusIssued).SetNotAfter(pem.NotAfter).SetNotBefore(pem.NotBefore).Save(ctx)
	if err != nil {
		return s.handleError("Error while collecting certificate", span, err)
	}
	return &pb.IssueSslResponse{Certificate: string(certificates.Certificate)}, nil
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
