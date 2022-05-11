package helper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
)

// LoadAccount loads an account from the given file
func LoadAccount(ctx context.Context, accountFile string, acmeDirectory string) (*acme.Client, error) {
	akey, err := LoadKeyFromPEMFile(accountFile, 0)
	if err != nil {
		return nil, err
	}

	client := &acme.Client{Key: akey, DirectoryURL: acmeDirectory}
	_, err = client.GetReg(ctx, "")

	return client, err
}

// RegisterAccount registers an account with the ACME server
func RegisterAccount(ctx context.Context, accountFile string, acmeDirectory string, eab acme.ExternalAccountBinding) (*acme.Client, error) {
	akey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	client := &acme.Client{Key: akey, DirectoryURL: acmeDirectory, KID: acme.KeyID(eab.KID)}
	_, err = client.Register(ctx, &acme.Account{ExternalAccountBinding: &eab}, acme.AcceptTOS)
	if err != nil {
		return nil, err
	}

	err = SaveToPEMFile(accountFile, akey, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type issueResult struct {
	certs [][]byte
	err   error
}

// RequestCertificate runs the acme flow to request a certificate with the desired contents
func RequestCertificate(ctx context.Context, span trace.Span, client *acme.Client, csr *x509.CertificateRequest, domains []string) ([][]byte, error) {

	span.AddEvent("Sending AuthorizeOrder Request")
	zap.L().Info("Sending AuthorizeOrder Request", zap.Strings("domains", domains))
	order, err := client.AuthorizeOrder(ctx, acme.DomainIDs(domains...))
	if err != nil {
		return nil, err
	}

	span.AddEvent("Received AuthorizeOrder Response. Requesting Certificate")
	zap.L().Info("Requesting certificate", zap.Strings("domains", domains))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	// Create a channel to received a signal that work is done.
	ch := make(chan issueResult, 1)
	go func() {
		crts, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr.Raw, true)
		// Report the work is done.
		ch <- issueResult{crts, err}
	}()
	select {
	case d := <-ch:
		if d.err != nil {
			return nil, d.err
		}
		span.AddEvent("Certificate received")
		zap.L().Info("Certificate received", zap.Strings("domains", domains))
		return d.certs, nil
	case <-ctx.Done():
		zap.L().Warn("Timeout waiting for certificate")
		span.AddEvent("Timeout waiting for certificate")
		span.SetStatus(codes.Error, "Timeout waiting for certificate")
		return nil, ctx.Err()
	}
}
