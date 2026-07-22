package grpc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"

	"github.com/hm-edu/harica/models"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/ent/predicate"
	"github.com/hm-edu/pki-service/pkg/acme"
	"github.com/hm-edu/pki-service/pkg/cfg"
	pkiHelper "github.com/hm-edu/pki-service/pkg/helper"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"

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
	ca := ""
	if x.Ca != nil {
		ca = *x.Ca
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
		Ca:                      ca,
		TransactionId:           x.TransactionId,
	}
}

func flattenCertificates(certs []*x509.Certificate) string {
	result := make([]byte, 0, len(certs))
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
	return nil, status.Error(codes.Internal, msg)
}

type sslAPIServer struct {
	pb.UnimplementedSSLServiceServer
	db     *ent.Client
	cfg    *cfg.PKIConfiguration
	logger *zap.Logger
	harica *haricaClients
	acme   *acme.Client

	last     *time.Time
	duration *time.Duration
}

func newSslAPIServer(cfg *cfg.PKIConfiguration, db *ent.Client, clients *haricaClients, acmeClient *acme.Client) *sslAPIServer {
	instance := &sslAPIServer{
		cfg:    cfg,
		logger: zap.L(),
		db:     db,
		harica: clients,
		acme:   acmeClient,
	}
	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "ssl_issue_last_duration",
		Help: "Required time for last SSL Certificates",
	}, func() float64 {
		if instance.duration != nil {
			return instance.duration.Seconds()
		}
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

	ca := "harica"
	if s.canUseAcme(csr, sans, logger) {
		ca = "letsencrypt"
	}

	logger = logger.With(zap.Strings("subject_alternative_names", sans), zap.String("ca", ca))
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
		SetCa(ca).
		AddDomainIDs(ids...).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while creating certificate", err, logger, hub)
	}

	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Requesting certificate", Category: "info", Level: sentry.LevelInfo}, nil)

	entry, err = s.db.Certificate.UpdateOneID(entry.ID).SetStatus(certificate.StatusRequested).Save(ctx)
	if err != nil {
		return s.handleError("Error while storing certificate", err, logger, hub)
	}

	if ca == "letsencrypt" {
		return s.issueAcmeCertificate(ctx, logger, hub, entry, csr, time.Now())
	}

	client, err := s.harica.Client()
	if err != nil {
		return s.handleError("Error while connecting to HARICA", err, logger, hub)
	}
	validationClient, err := s.harica.Validation()
	if err != nil {
		return s.handleError("Error while connecting to HARICA", err, logger, hub)
	}

	// Check which organization the domains belong to
	orgs, err := retryHarica(ctx, logger, client, "CheckMatchingOrganization", func() ([]models.OrganizationResponse, error) {
		return client.CheckMatchingOrganization(sans)
	})
	if err != nil || len(orgs) == 0 {
		return s.handleError("Error while checking organization", err, logger, hub)
	}

	transaction, err := retryHarica(ctx, logger, client, "RequestCertificate", func() (*models.CertificateRequestResponse, error) {
		return client.RequestCertificate(sans, req.Csr, s.cfg.CertType, orgs[0])
	})
	if err != nil {
		return s.handleError("Error while requesting certificate", err, logger, hub)
	}

	entry, err = s.db.Certificate.UpdateOneID(entry.ID).SetTransactionId(transaction.TransactionID).Save(ctx)
	if err != nil {
		return s.handleError("Error while storing certificate", err, logger, hub)
	}

	reviews, err := retryHarica(ctx, logger, validationClient, "GetPendingReviews", func() ([]models.ReviewResponse, error) {
		return validationClient.GetPendingReviews()
	})
	if err != nil {
		return s.handleError("Error while fetching pending reviews", err, logger, hub)
	}

	logger.Info("Certificate requested. Approving Request", zap.String("transaction_id", transaction.TransactionID))
	for _, r := range reviews {
		if r.TransactionID == transaction.TransactionID {
			for _, sub := range r.ReviewGetDTOs {
				err = retryHaricaVoid(ctx, logger, validationClient, "ApproveRequest", func() error {
					return validationClient.ApproveRequest(sub.ReviewID, "Auto Approval", sub.ReviewValue)
				})
				if err != nil {
					return s.handleError("Error while approving request", err, logger, hub)
				}
			}
			break
		}
	}

	if !req.WaitForIssue {
		logger.Info("Request approved. Certificate will be collected later", zap.String("transaction_id", transaction.TransactionID))
		return &pb.IssueSslResponse{TransactionId: transaction.TransactionID}, nil
	}
	logger.Info("Request approved. Collecting certificate")
	cert, err := retryHarica(ctx, logger, client, "GetCertificate", func() (*models.CertificateResponse, error) {
		return client.GetCertificate(transaction.TransactionID)
	})
	if err != nil {
		return s.handleError("Error while obtaining certificate", err, logger, hub)
	}
	logger.Info("Certificate collected")
	return s.storeCollectedCertificate(ctx, logger, hub, entry, transaction.TransactionID, cert)
}

// CollectCertificate tries to collect a certificate that was requested via
// IssueCertificate with wait_for_issue disabled. Every call performs at most
// one collection attempt against HARICA; as long as the certificate has not
// been issued yet, a response without certificate is returned so that the
// caller can poll again with a fresh, short-lived request.
func (s *sslAPIServer) CollectCertificate(ctx context.Context, req *pb.CollectSslRequest) (*pb.IssueSslResponse, error) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	log := s.logger
	if hub != nil && hub.Scope() != nil {
		log = log.With(zapsentry.NewScopeFromScope(hub.Scope()))
	}
	logger := log.With(zap.String("transaction_id", req.TransactionId))

	if req.TransactionId == "" {
		return nil, status.Error(codes.InvalidArgument, "No transaction id provided")
	}

	entry, err := s.db.Certificate.Query().Where(certificate.TransactionId(req.TransactionId)).First(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Certificate not found")
	}

	client, err := s.harica.Client()
	if err != nil {
		return s.handleError("Error while connecting to HARICA", err, logger, hub)
	}
	validationClient, err := s.harica.Validation()
	if err != nil {
		return s.handleError("Error while connecting to HARICA", err, logger, hub)
	}

	// The pending review is sometimes not visible yet while the certificate
	// is requested. Re-check the reviews on every poll until the certificate
	// is issued so a late review does not stall the transaction forever.
	if entry.Status != certificate.StatusIssued {
		reviews, err := runHaricaOnce(validationClient, func() ([]models.ReviewResponse, error) {
			return validationClient.GetPendingReviews()
		})
		if err != nil {
			logger.Warn("Fetching pending reviews failed", zap.Error(err))
			if isAuthError(err) {
				_ = validationClient.SessionRefresh(true)
			}
		} else {
			for _, r := range reviews {
				if r.TransactionID != req.TransactionId {
					continue
				}
				logger.Info("Approving pending request")
				for _, sub := range r.ReviewGetDTOs {
					if _, err := runHaricaOnce(validationClient, func() (struct{}, error) {
						return struct{}{}, validationClient.ApproveRequest(sub.ReviewID, "Auto Approval", sub.ReviewValue)
					}); err != nil {
						logger.Warn("Approving request failed", zap.Error(err))
					}
				}
				break
			}
		}
	}

	cert, err := runHaricaOnce(client, func() (*models.CertificateResponse, error) {
		return client.GetCertificate(req.TransactionId)
	})
	if err != nil {
		if isCertificatePending(err) {
			logger.Info("Certificate not issued yet")
			return &pb.IssueSslResponse{TransactionId: req.TransactionId}, nil
		}
		if isAuthError(err) {
			_ = client.SessionRefresh(true)
		}
		logger.Warn("Collecting certificate failed", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Collecting certificate failed")
	}
	if cert.PemBundle == "" {
		logger.Info("Certificate not issued yet")
		return &pb.IssueSslResponse{TransactionId: req.TransactionId}, nil
	}
	logger.Info("Certificate collected")
	return s.storeCollectedCertificate(ctx, logger, hub, entry, req.TransactionId, cert)
}

// storeCollectedCertificate parses a certificate bundle collected from HARICA,
// persists the certificate metadata and returns the certificate chain.
func (s *sslAPIServer) storeCollectedCertificate(ctx context.Context, logger *zap.Logger, hub *sentry.Hub, entry *ent.Certificate, transactionID string, cert *models.CertificateResponse) (*pb.IssueSslResponse, error) {
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Certificate collected", Category: "info", Level: sentry.LevelInfo}, nil)
	certs, err := pkiHelper.ParseCertificates([]byte(cert.PemBundle))
	if err != nil {
		return s.handleError("Error parsing certificate", err, logger, hub)
	}
	stop := time.Now()
	duration := stop.Sub(entry.CreateTime)
	s.duration = &duration
	s.last = &stop
	pem := certs[0]
	serial := fmt.Sprintf("%032x", pem.SerialNumber)
	logger.Info("Certificate issued",
		zap.Duration("duration", duration),
		zap.String("serial", serial))

	_, err = s.db.Certificate.UpdateOneID(entry.ID).
		SetSerial(pkiHelper.NormalizeSerial(serial)).
		SetStatus(certificate.StatusIssued).
		SetNotAfter(pem.NotAfter).
		SetNotBefore(pem.NotBefore).
		SetCreated(stop).
		SetTransactionId(transactionID).
		Save(ctx)

	if err != nil {
		return s.handleError("Error while saving collected certificate", err, logger, hub)
	}

	return &pb.IssueSslResponse{Certificate: flattenCertificates(certs), TransactionId: transactionID}, nil
}

// canUseAcme reports whether the requested certificate can be issued by the
// ACME CA. The ACME order is derived from the CSR, so the CSR must contain
// exactly the requested domains and all of them must be covered by the DNS
// validation config. Requests that do not qualify fall back to HARICA so
// zones can be migrated one by one.
func (s *sslAPIServer) canUseAcme(csr *x509.CertificateRequest, sans []string, logger *zap.Logger) bool {
	if s.acme == nil {
		return false
	}
	csrDomains := make(map[string]bool)
	if csr.Subject.CommonName != "" {
		csrDomains[strings.ToLower(csr.Subject.CommonName)] = true
	}
	for _, d := range csr.DNSNames {
		csrDomains[strings.ToLower(d)] = true
	}
	requested := make(map[string]bool)
	for _, san := range sans {
		requested[strings.ToLower(san)] = true
	}
	for san := range requested {
		if !csrDomains[san] {
			logger.Warn("Domain missing in CSR, falling back to HARICA", zap.String("domain", san))
			return false
		}
	}
	for d := range csrDomains {
		if !requested[d] {
			logger.Warn("CSR contains additional domain, falling back to HARICA", zap.String("domain", d))
			return false
		}
	}
	if !s.acme.Covers(sans) {
		logger.Info("Domains not covered by DNS validation config, falling back to HARICA")
		return false
	}
	return true
}

func (s *sslAPIServer) issueAcmeCertificate(ctx context.Context, logger *zap.Logger, hub *sentry.Hub, entry *ent.Certificate, csr *x509.CertificateRequest, start time.Time) (*pb.IssueSslResponse, error) {
	certPEM, err := s.acme.ObtainForCSR(ctx, csr)
	if err != nil {
		return s.handleError("Error while requesting certificate", err, logger, hub)
	}
	hub.AddBreadcrumb(&sentry.Breadcrumb{Message: "Certificate collected", Category: "info", Level: sentry.LevelInfo}, nil)
	stop := time.Now()
	duration := stop.Sub(start)
	s.duration = &duration
	s.last = &stop
	certs, err := pkiHelper.ParseCertificates(certPEM)
	if err != nil {
		return s.handleError("Error parsing certificate", err, logger, hub)
	}
	leaf := certs[0]
	serial := fmt.Sprintf("%032x", leaf.SerialNumber)
	logger.Info("Certificate issued",
		zap.Duration("duration", duration),
		zap.String("serial", serial))

	_, err = s.db.Certificate.UpdateOneID(entry.ID).
		SetSerial(pkiHelper.NormalizeSerial(serial)).
		SetStatus(certificate.StatusIssued).
		SetNotAfter(leaf.NotAfter).
		SetNotBefore(leaf.NotBefore).
		SetCreated(stop).
		SetCertificate(string(certPEM)).
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

	// The HARICA revocation reason is only required (and fetched) if a
	// HARICA certificate is actually revoked.
	getReason := sync.OnceValues(func() (*models.RevocationReasonsResponse, error) {
		client, err := s.harica.Client()
		if err != nil {
			return nil, err
		}
		reasons, err := retryHarica(ctx, logger, client, "GetRevocationReasons", func() ([]models.RevocationReasonsResponse, error) {
			return client.GetRevocationReasons()
		})
		if err != nil {
			return nil, err
		}
		for _, r := range reasons {
			if r.Name == "4.9.1.1.1.1" {
				return &r, nil
			}
		}
		return nil, fmt.Errorf("revocation reason not found")
	})

	// revokeOne revokes a single certificate using the CA it was issued by.
	// Certificates from unknown CAs are skipped without an error.
	revokeOne := func(c *ent.Certificate, logger *zap.Logger) error {
		ca := ""
		if c.Ca != nil {
			ca = *c.Ca
		}
		switch ca {
		case "letsencrypt":
			if s.acme == nil {
				return fmt.Errorf("certificate %d was issued via ACME but no ACME client is configured", c.ID)
			}
			if c.Certificate == nil || *c.Certificate == "" {
				logger.Warn("Certificate has no stored PEM. Cannot revoke", zap.Int("id", c.ID))
				return nil
			}
			logger.Info("Revoking certificate via ACME", zap.Int("id", c.ID))
			if err := s.acme.Revoke(ctx, []byte(*c.Certificate)); err != nil {
				return err
			}
		case "harica":
			if c.TransactionId == "" {
				logger.Warn("Certificate has no transaction id", zap.Int("id", c.ID))
				return nil
			}
			validationClient, err := s.harica.Validation()
			if err != nil {
				return err
			}
			reason, err := getReason()
			if err != nil {
				return err
			}
			logger.Info("Revoking certificate", zap.String("transaction_id", c.TransactionId), zap.String("reason", reason.Name), zap.String("description", req.Reason))
			err = retryHaricaVoid(ctx, logger, validationClient, "RevokeCertificate", func() error {
				return validationClient.RevokeCertificate(*reason, "", c.TransactionId)
			})
			if err != nil {
				return err
			}
		default:
			logger.Info("Skipping certificate. Not issued by HARICA or ACME", zap.Int("id", c.ID))
			return nil
		}
		_, err := s.db.Certificate.UpdateOneID(c.ID).SetStatus(certificate.StatusRevoked).Save(ctx)
		return err
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
		if err := revokeOne(c, logger); err != nil {
			logger.Error("Revoking request failed", zap.Error(err))
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
				ret <- struct{ err error }{revokeOne(c, logger)}
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
