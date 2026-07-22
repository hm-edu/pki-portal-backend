package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	legoacme "github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"go.uber.org/zap"
)

// account implements the registration.User interface of lego.
type account struct {
	email        string
	registration *legoacme.ExtendedAccount
	key          crypto.Signer
}

func (a *account) GetEmail() string                           { return a.email }
func (a *account) GetRegistration() *legoacme.ExtendedAccount { return a.registration }
func (a *account) GetPrivateKey() crypto.Signer               { return a.key }

// Client wraps a lego ACME client that validates domains using DNS-01
// challenges published via RFC2136/TSIG. The ACME session (account key and
// registration) is created once and reused for all requests.
type Client struct {
	lego   *lego.Client
	dns    *DNSConfig
	logger *zap.Logger
}

// NewClient creates a new ACME client. The account key is loaded from
// keyPath; if the file does not exist, a new key is generated, stored there
// and a new ACME account is registered.
func NewClient(ctx context.Context, email, directory, keyPath string, dnsCfg *DNSConfig, logger *zap.Logger) (*Client, error) {
	if email == "" {
		return nil, errors.New("no ACME account email configured")
	}
	key, created, err := loadOrCreateKey(keyPath)
	if err != nil {
		return nil, err
	}
	acc := &account{email: email, key: key}
	cfg := lego.NewConfig(acc)
	cfg.CADirURL = directory
	client, err := lego.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating ACME client: %w", err)
	}
	// Use the public DNS view for propagation checks and authoritative
	// nameserver discovery. The system resolver may expose an internal view in
	// split-DNS environments that is not visible to the ACME CA.
	dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{
		RecursiveNameservers: []string{"1.1.1.1:53", "8.8.8.8:53"},
	}))
	if err := client.Challenge.SetDNS01Provider(NewDNSProvider(dnsCfg, logger)); err != nil {
		return nil, fmt.Errorf("setting DNS-01 provider: %w", err)
	}

	reg, err := client.Registration.ResolveAccountByKey(ctx)
	if err != nil {
		if !created {
			logger.Warn("ACME account not found for existing key, registering new account", zap.Error(err))
		}
		reg, err = client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, fmt.Errorf("registering ACME account: %w", err)
		}
	}
	acc.registration = reg
	logger.Info("ACME account ready", zap.String("email", email), zap.String("directory", directory))

	return &Client{lego: client, dns: dnsCfg, logger: logger}, nil
}

// Covers reports whether all given domains can be validated with the
// configured DNS zones.
func (c *Client) Covers(domains []string) bool {
	return c.dns.Covers(domains)
}

// ObtainForCSR requests a certificate for the given CSR. The returned bytes
// contain the full PEM encoded chain (leaf first).
func (c *Client) ObtainForCSR(ctx context.Context, csr *x509.CertificateRequest) ([]byte, error) {
	res, err := c.lego.Certificate.ObtainForCSR(ctx, certificate.ObtainForCSRRequest{
		CSR:    csr,
		Bundle: true,
	})
	if err != nil {
		return nil, err
	}
	return res.Certificate, nil
}

// Revoke revokes the given PEM encoded certificate. Revoking an already
// revoked certificate is not treated as an error.
func (c *Client) Revoke(ctx context.Context, certPEM []byte) error {
	err := c.lego.Certificate.Revoke(ctx, certPEM)
	if err == nil {
		return nil
	}
	var problem *legoacme.ProblemDetails
	if errors.As(err, &problem) && problem.Type == legoacme.AlreadyRevokedErrorType {
		c.logger.Info("Certificate was already revoked")
		return nil
	}
	return err
}

// loadOrCreateKey loads the ACME account key from the given path or creates
// a new one if the file does not exist. It reports whether a new key was
// created.
func loadOrCreateKey(path string) (crypto.Signer, bool, error) {
	if path == "" {
		return nil, false, errors.New("no ACME account key path configured")
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is provided by the operator
	if err == nil {
		key, err := certcrypto.ParsePEMPrivateKey(data)
		if err != nil {
			return nil, false, fmt.Errorf("parsing ACME account key %s: %w", path, err)
		}
		return key, false, nil
	}
	if !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("reading ACME account key %s: %w", path, err)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, false, err
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, false, err
	}
	data = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, false, fmt.Errorf("storing ACME account key %s: %w", path, err)
	}
	return key, true, nil
}
