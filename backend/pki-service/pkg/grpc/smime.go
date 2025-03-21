package grpc

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/go-co-op/gocron"
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
	haricaClient, err := harica.NewClient(
		cfg.ValidationUser,
		cfg.ValidationPassword,
		cfg.ValidationTotpSeed,
		harica.WithRefreshInterval(5*time.Minute),
		harica.WithRetry(3),
	)
	if err != nil {
		return nil, err
	}
	instance := &smimeAPIServer{
		cfg:              cfg,
		logger:           zap.L(),
		validationClient: haricaClient,
		db:               db,
	}
	s := gocron.NewScheduler(time.UTC)
	_, err = s.Every(1).Hour().Do(func() {
		err = haricaClient.SessionRefresh(true)
		if err != nil {
			instance.logger.Error("Error refreshing validation client", zap.Error(err))
		}
	})
	if err != nil {
		return nil, err
	}
	s.StartAsync()
	return instance, nil
}

const (
	geantIssuer = `-----BEGIN CERTIFICATE-----
MIIGRDCCBCygAwIBAgIQFfmubKqNLtTTb3h/Htx7ATANBgkqhkiG9w0BAQsFADBv
MQswCQYDVQQGEwJHUjE3MDUGA1UECgwuSGVsbGVuaWMgQWNhZGVtaWMgYW5kIFJl
c2VhcmNoIEluc3RpdHV0aW9ucyBDQTEnMCUGA1UEAwweSEFSSUNBIENsaWVudCBS
U0EgUm9vdCBDQSAyMDIxMB4XDTI1MDEwMzExMTMwOFoXDTM5MTIzMTExMTMwN1ow
YzELMAkGA1UEBhMCR1IxNzA1BgNVBAoMLkhlbGxlbmljIEFjYWRlbWljIGFuZCBS
ZXNlYXJjaCBJbnN0aXR1dGlvbnMgQ0ExGzAZBgNVBAMMEkdFQU5UIFMvTUlNRSBS
U0EgMTCCAaIwDQYJKoZIhvcNAQEBBQADggGPADCCAYoCggGBAKu4bq/+byKjHo25
Xz32YBmO+Wrkmc+UmfcdXSCI7yawwU9JSMEHAAKAASaJpLr9JAyt+tlB/rn/Sazn
SwY4ipBIffR0D5k/ndfiI553dWgI4i/tkOGlNej/7JyE2CS9kTlOOs6pg5HaDpwq
jAhCkje+IByg5gKWH6lzvMJo5jQOtsGB2q6e5cYKwa9LJOAcR8iquds9LFssbHSM
uVdSuTjpAjcGLqWfW++C0YXpWD+UonjQ6lNEuiKUDmrFc+SEtLw56lYtp4uuxm4L
W/HQSsx+oGwMBqaR6HhBQ3LydONjsbcbegRqJZFJoLsnwIHorEag44UIvjXzYJAx
/NTiwVdHldO7cEvWscDbyQLR9koBoliq2HrgYFQs7NQxU+7MLNSh8i6znWVNISUE
g36M//I8BZl4VqD70ELlhKKN7rx+i7BwKOd2gxdWgFJhkPyQu9o+82R9epXiRblo
/rdkyv+2BFR7VpbgPUzncdi8/0h4dP/qQFYnA+Df0FFj7gYczwIDAQABo4IBZjCC
AWIwEgYDVR0TAQH/BAgwBgEB/wIBADAfBgNVHSMEGDAWgBSg1gc9XiT3e6BELiRS
DRmqKwSRpzBQBggrBgEFBQcBAQREMEIwQAYIKwYBBQUHMAKGNGh0dHA6Ly9jcnQu
aGFyaWNhLmdyL0hBUklDQS1DbGllbnQtUm9vdC0yMDIxLVJTQS5jZXIwRAYDVR0g
BD0wOzA5BgRVHSAAMDEwLwYIKwYBBQUHAgEWI2h0dHA6Ly9yZXBvLmhhcmljYS5n
ci9kb2N1bWVudHMvQ1BTMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDBDBF
BgNVHR8EPjA8MDqgOKA2hjRodHRwOi8vY3JsLmhhcmljYS5nci9IQVJJQ0EtQ2xp
ZW50LVJvb3QtMjAyMS1SU0EuY3JsMB0GA1UdDgQWBBTrsi87/a4CzCpEBl0lzR0S
ImiwRzAOBgNVHQ8BAf8EBAMCAYYwDQYJKoZIhvcNAQELBQADggIBADveuEX23Dwr
kygKtsF7DmcTGmi8SE20jmJLe0TMT8Nws1NqppE0ACym1agtY1IjUFm5MWabG/Ic
vRTh8sB9cRZgDQMqZLNCLofqL4aj/dKBXH4bwH2MVdjNHBoGvZkyhRz/kBE+x1va
WXclhWQMOX5nVvRMfiEJiYotMP7KM88IaVZ9DkGJJEVftsnUWuvCWUtjagD6XWlq
LHjNl+LufiZ/h9lDvaWqG1/obfdStgofMc30RL+ES6gYKRwZpCA1coFzXV7Cnwx8
toTl8bReqCNXexKzxlqAcRXPOmlKkJQuqRI297oNuMPnoNZCY+yLnxyd4kZuu0Xc
OTNTpVjM8bvg8ACqhSYanrNDi/zTiTk7gwm9GyH1X45fFNGNEFgpIaApjT2UELuk
DOmP18ZwC4EQeHawPJIqffMEmUJm6qbRPKGnNmcyygh4iZU3QbkRLLp3Z6QV3WoT
Eqyf5mL9qTGS6WJG65L8oaKw1Xh/bdGuVIDyBahpfP2c2pCd0UH6+x73Rrq9GFlO
ijVr2OQSvKhzETNG917SvcURCBhMnIQFUXqHQyIY60eH1po6WtNOq/1K5kpOG6Sq
1RVc02LEit48uK4tRMVUKekSOjruGXW38DmAriPcMHjI6VQbqjc0Sq1VPz76ee4F
M5uLviSUZHYqDDqMWa8LFImK9iiKI8E3
-----END CERTIFICATE-----`
	haricaRoot = `-----BEGIN CERTIFICATE-----
MIIFqjCCA5KgAwIBAgIQVVL4HtsbJCyeu5YYzQIoPjANBgkqhkiG9w0BAQsFADBv
MQswCQYDVQQGEwJHUjE3MDUGA1UECgwuSGVsbGVuaWMgQWNhZGVtaWMgYW5kIFJl
c2VhcmNoIEluc3RpdHV0aW9ucyBDQTEnMCUGA1UEAwweSEFSSUNBIENsaWVudCBS
U0EgUm9vdCBDQSAyMDIxMB4XDTIxMDIxOTEwNTg0NloXDTQ1MDIxMzEwNTg0NVow
bzELMAkGA1UEBhMCR1IxNzA1BgNVBAoMLkhlbGxlbmljIEFjYWRlbWljIGFuZCBS
ZXNlYXJjaCBJbnN0aXR1dGlvbnMgQ0ExJzAlBgNVBAMMHkhBUklDQSBDbGllbnQg
UlNBIFJvb3QgQ0EgMjAyMTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIB
AIHbV0KQLHQ19Pi4dBlNqwlad0WBc2KwNZ/40LczAIcTtparDlQSMAe8m7dI19EZ
g66O2KnxqQCEsIxenugMj1Rpv/bUCE8mcP4YQWMaszKLQPgHq1cx8MYWdmeatN0v
8tFrxdCShJFxbg8uY+kfU6TdUhPMCYMpgQzFU3VEsQ5nUxjQwx+IS5+UJLQpvLvo
Tv1v0hUdSdyNcPIRGiBRVRG6iG/E91B51qox4oQ9XjLIdypQceULL+m26u+rCjM5
Dv2PpWdDgo6YaQkJG0DNOGdH6snsl3ES3iT1cjzR90NMJveQsonpRUtVPTEFekHi
lbpDwBfFtoU9GY1kcPNbrM2f0yl1h0uVZ2qm+NHdvJCGiUMpqTdb9V2wJlpTQnaQ
K8+eVmwrVM9cmmXfW4tIYDh8+8ULz3YEYwIzKn31g2fn+sZD/SsP1CYvd6QywSTq
ZJ2/szhxMUTyR7iiZkGh+5t7vMdGanW/WqKM6GpEwbiWtcAyCC17dDVzssrG/q8R
chj258jCz6Uq6nvWWeh8oLJqQAlpDqWW29EAufGIbjbwiLKd8VLyw3y/MIk8Cmn5
IqRl4ZvgdMaxhZeWLK6Uj1CmORIfvkfygXjTdTaefVogl+JSrpmfxnybZvP+2M/u
vZcGHS2F3D42U5Z7ILroyOGtlmI+EXyzAISep0xxq0o3AgMBAAGjQjBAMA8GA1Ud
EwEB/wQFMAMBAf8wHQYDVR0OBBYEFKDWBz1eJPd7oEQuJFINGaorBJGnMA4GA1Ud
DwEB/wQEAwIBhjANBgkqhkiG9w0BAQsFAAOCAgEADUf5CWYxUux57sKo8mg+7ZZF
yzqmmGM/6itNTgPQHILhy9Pl1qtbZyi8nf4MmQqAVafOGyNhDbBX8P7gyr7mkNuD
LL6DjvR5tv7QDUKnWB9p6oH1BaX+RmjrbHjJ4Orn5t4xxdLVLIJjKJ1dqBp+iObn
K/Es1dAFntwtvTdm1ASip62/OsKoO63/jZ0z4LmahKGHH3b0gnTXDvkwSD5biD6q
XGvWLwzojnPCGJGDObZmWtAfYCddTeP2Og1mUJx4e6vzExCuDy+r6GSzGCCdRjVk
JXPqmxBcWDWJsUZIp/Ss1B2eW8yppRoTTyRQqtkbbbFA+53dWHTEwm8UcuzbNZ+4
VHVFw6bIGig1Oq5l8qmYzq9byTiMMTt/zNyW/eJb1tBZ9Ha6C8tPgxDHQNAdYOkq
5UhYdwxFab4ZcQQk4uMkH0rIwT6Z9ZaYOEgloRWwG9fihBhb9nE1mmh7QMwYXAwk
ndSV9ZmqRuqurL/0FBkk6Izs4/W8BmiKKgwFXwqXdafcfsD913oY3zDROEsfsJhw
v8x8c/BuxDGlpJcdrL/ObCFKvicjZ/MGVoEKkY624QMFMyzaNAhNTlAjrR+lxdR6
/uoJ7KcoYItGfLXqm91P+edrFcaIz0Pb5SfcBFZub0YV8VYt6FwMc8MjgTggy8kM
ac8sqzuEYDMZUv1pFDM=
-----END CERTIFICATE-----`
)

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
	params := models.SmimeBulkRequest{
		Email:        req.Email,
		FriendlyName: req.CommonName,
		CertType:     "email_only",
		CSR:          req.Csr,
	}
	if ok, err := regexp.MatchString(`^[a-zA-Z \-]`, req.FirstName); ok && err == nil {
		if ok, err := regexp.MatchString(`^[a-zA-Z \-]`, req.LastName); ok && err == nil {
			params.GivenName = req.FirstName
			params.Surname = req.LastName
			params.CertType = "natural_legal_lcp"
		}
	}
	cert, err := s.validationClient.RequestSmimeBulkCertificates(groups[0].OrganizationID, params)
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

	_, err = s.db.SmimeCertificate.Create().
		SetCreateTime(time.Now()).
		SetEmail(req.Email).
		SetSerial(certX509.SerialNumber.String()).
		SetNotAfter(certX509.NotAfter).
		SetNotBefore(certX509.NotBefore).
		SetStatus(smimecertificate.StatusIssued).
		SetTransactionId(cert.TransactionID).Save(ctx)

	if err != nil {
		hub.CaptureException(err)
		logger.Error("Error saving certificate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error saving certificate")
	}

	cert.Certificate = fmt.Sprintf("%s\n%s\n%s",
		cert.Certificate,
		geantIssuer,
		haricaRoot,
	)

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

	reasons, err := s.validationClient.GetRevocationReasons()
	if err != nil {
		return nil, status.Error(codes.Internal, "Error fetching revocation reasons")
	}
	var reason *models.RevocationReasonsResponse
	for _, r := range reasons {
		if r.Name == "4.9.1.1.1.1" {
			reason = &r
			break
		}
	}
	if reason == nil {
		return nil, status.Error(codes.Internal, "Error fetching revocation reasons")
	}

	switch req.Identifier.(type) {
	case *pb.RevokeSmimeRequest_Email:
		logger = logger.With(zap.String("email", req.GetEmail()))
		logger.Info("Revoking smime certificate")
		certs, err := s.db.SmimeCertificate.Query().Where(smimecertificate.EmailEQ(req.GetEmail())).All(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error fetching certificates")
		}
		if len(certs) == 0 {
			return nil, status.Error(codes.NotFound, "Certificate not found")
		}
		for _, cert := range certs {
			s.logger.Info("Revoking smime certificate", zap.String("serial", req.GetSerial()), zap.String("email", cert.Email), zap.String("transaction_id", cert.TransactionId))
			err := s.validationClient.RevokeSmimeBulkCertificateEntry(cert.TransactionId, req.Reason, reason.Name)
			if err != nil {
				return nil, status.Error(codes.Internal, "Error revoking certificate")
			}
			_, err = s.db.SmimeCertificate.UpdateOneID(cert.ID).SetStatus(smimecertificate.StatusRevoked).Save(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, "Error updating certificate")
			}

		}

		logger.Info("Successfully revoked smime certificate")
		return &emptypb.Empty{}, nil
	case *pb.RevokeSmimeRequest_Serial:
		certs, err := s.db.SmimeCertificate.Query().Where(smimecertificate.Serial(req.GetSerial())).All(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error fetching certificates")
		}
		if len(certs) == 0 {
			return nil, status.Error(codes.NotFound, "Certificate not found")
		}
		for _, cert := range certs {
			s.logger.Info("Revoking smime certificate", zap.String("serial", req.GetSerial()), zap.String("email", cert.Email), zap.String("transaction_id", cert.TransactionId))
			err := s.validationClient.RevokeSmimeBulkCertificateEntry(cert.TransactionId, req.Reason, reason.Name)
			if err != nil {
				return nil, status.Error(codes.Internal, "Error revoking certificate")
			}
			_, err = s.db.SmimeCertificate.UpdateOneID(cert.ID).SetStatus(smimecertificate.StatusRevoked).Save(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, "Error updating certificate")
			}
		}

		return &emptypb.Empty{}, nil
	}

	return nil, status.Errorf(codes.Unimplemented, "method RevokeCertificate not implemented")
}
