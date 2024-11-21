package grpc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	legoCert "github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	legoLog "github.com/go-acme/lego/v4/log"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/ent/predicate"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pkiHelper "github.com/hm-edu/pki-service/pkg/helper"
	"github.com/hm-edu/pki-service/pkg/helper/precheck"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

func (s *sslAPIServer) handleError(msg string, err error, logger *zap.Logger, hub *sentry.Hub) (*pb.IssueSslResponse, error) {
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: msg, Category: "error", Level: sentry.LevelError}, nil)
	hub.CaptureException(err)
	logger.Error(msg, zap.Error(err))
	return nil, status.Errorf(codes.Internal, msg)
}

type sslAPIServer struct {
	pb.UnimplementedSSLServiceServer
	client     *sectigo.Client
	db         *ent.Client
	cfg        *cfg.SectigoConfiguration
	logger     *zap.Logger
	legoClient *lego.Client

	last     *time.Time
	duration *time.Duration
}

func newSslAPIServer(client *sectigo.Client, cfg *cfg.SectigoConfiguration, db *ent.Client) *sslAPIServer {
	legoClient := registerAcme(cfg)
	instance := &sslAPIServer{client: client, legoClient: legoClient, cfg: cfg, logger: zap.L(), db: db}
	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "ssl_issue_last_duration",
		Help: "Required time for last SSL Certificates",
	}, func() float64 {
		if instance.duration != nil {
			return instance.duration.Seconds()
		return 0
	})

	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "ssl_issue_last_timestamp",
		Help: "Timestamp of last SSL Certificate",
	}, func() float64 {
		if instance.last != nil {
			return float64(instance.last.UnixMilli())
		}
		return 0
	})

	return instance
}

func registerAcme(cfg *cfg.SectigoConfiguration) *lego.Client {
	accountFile := filepath.Join(cfg.AcmeStorage, "reg.json")
	keyFile := filepath.Join(cfg.AcmeStorage, "reg.key")

	var account pkiHelper.User
	if ok, _ := pkiHelper.FileExists(accountFile); !ok {
		// Actually we would not need a private key but the lego API requires one.
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil
		}

		account = pkiHelper.User{
			Key: privateKey, Email: cfg.AcmeEmail,
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
		account.Key, err = pkiHelper.LoadPrivateKey(keyFile)
		if err != nil {
			return nil
		}

	}
	legoCfg := lego.NewConfig(&account)

	legoLog.Logger = pkiHelper.NewZapLogger(zap.L())
	legoCfg.Certificate.Timeout = time.Duration(5) * time.Minute
	if account.Registration == nil {
		legoClient, err := lego.NewClient(legoCfg)

		if err != nil {
			return nil
		}
		err = pkiHelper.RegisterAcme(legoClient, account, accountFile, keyFile)
		if err != nil {
			return nil
		}
	}

	legoClient, err := lego.NewClient(legoCfg)
	if err != nil {
		zap.L().Fatal("Failed to create lego client", zap.Error(err))
	}
	dns, err := pkiHelper.NewDNSProvider(cfg.DnsConfigs)
	if err != nil {
		zap.L().Fatal("Failed to create DNS provider", zap.Error(err))
	}

	err = legoClient.Challenge.SetDNS01Provider(dns, dns01.WrapPreCheck(precheck.CheckDNS))
	if err != nil {
		zap.L().Fatal("Failed to set DNS01 provider", zap.Error(err))
	}

	return legoClient
}

func (s *sslAPIServer) CertificateDetails(ctx context.Context, req *pb.CertificateDetailsRequest) (*pb.SslCertificateDetails, error) {
	x, err := s.db.Certificate.Query().WithDomains().Where(certificate.Serial(req.Serial)).First(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Certificate not found")
	}
	return mapCertificate(x), nil
}

func (s *sslAPIServer) ListCertificates(ctx context.Context, req *pb.ListSslRequest) (*pb.ListSslResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	logger := log.With(zap.Strings("domains", req.Domains), zap.Bool("partial", req.IncludePartial))
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
		return nil, status.Error(codes.Internal, "Error querying certificates")
	}
	return &pb.ListSslResponse{Items: helper.Map(certificates, mapCertificate)}, nil
}

func (s *sslAPIServer) IssueCertificate(ctx context.Context, req *pb.IssueSslRequest) (*pb.IssueSslResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
		hub.Scope().SetUser(sentry.User{Email: req.Issuer})
	}

	logger := log.With(zap.String("issuer", req.Issuer))

	block, _ := pem.Decode([]byte(req.Csr))

	if block == nil {
		hub.CaptureMessage("Invalid pem block")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid pem block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Error parsing CSR", Category: "error", Level: sentry.LevelError}, nil)
		hub.CaptureException(err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid CSR")
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
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

	logger = logger.With(zap.Strings("subject_alternative_names", sans))
	logger.Info("Issuing new server certificate")

	for _, fqdn := range sans {
		id, err := s.db.Domain.Create().
			SetFqdn(fqdn).
			OnConflictColumns(domain.FieldFqdn).
			Ignore().
			ID(ctx)

		if err != nil {
			return s.handleError("Error while creating certificate", err, logger, hub)
		}
		ids = append(ids, id)
	}

	entry, err := s.db.Certificate.Create().
		SetCommonName(sans[0]).
		SetIssuedBy(req.Issuer).
		SetSource(req.Source).
		SetCa("zerossl").
		AddDomainIDs(ids...).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while creating certificate", err, logger, hub)
	}

	start := time.Now()
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Requesting certificate", Category: "info", Level: sentry.LevelInfo}, nil)

	entry, err = s.db.Certificate.UpdateOneID(entry.ID).SetStatus(certificate.StatusRequested).Save(ctx)
	if err != nil {
		return s.handleError("Error while storing certificate", err, logger, hub)
	}

	resp, err := s.legoClient.Certificate.ObtainForCSR(legoCert.ObtainForCSRRequest{CSR: csr, Bundle: true})
	if err != nil {
		return s.handleError("Error while obtaining certificate", err, logger)
	}
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Certificate collected", Category: "info", Level: sentry.LevelInfo}, nil)
	stop := time.Now()
	duration := stop.Sub(start)
	s.duration = &duration
	s.last = &stop

	certs, err := pkiHelper.ParseCertificates(resp.Certificate)
	if err != nil {
		return s.handleError("Error parsing certificate", err, logger, hub)
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
		return s.handleError("Error while saving collected certificate", err, logger, hub)
	}

	return &pb.IssueSslResponse{Certificate: flattenCertificates(certs)}, nil
}

func (s *sslAPIServer) RevokeCertificate(ctx context.Context, req *pb.RevokeSslRequest) (*emptypb.Empty, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}

	logger := log.With(zap.String("reason", req.Reason))

	errorReturn := func(err error, logger *zap.Logger) (*emptypb.Empty, error) {
		logger.Error("Failed to revoke certificate", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to revoke certificate")
	}

	switch req.Identifier.(type) {
	case *pb.RevokeSslRequest_Serial:
		serial := pkiHelper.NormalizeSerial(req.GetSerial())
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
