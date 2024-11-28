package helper

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/hm-edu/pki-service/pkg/cfg"
)

// User represents an ACME user.
type User struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

// GetEmail returns the email of the user.
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the registration resource of the user.
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the private key of the user.
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

// RegisterAcme performs a new registration and stores the registration in the given file.
func RegisterAcme(client *lego.Client, config *cfg.PKIConfiguration, account User, accountFile string, keyFile string) error {
	reg, err := client.Registration.RegisterWithExternalAccountBinding(
		registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  config.AcmeKid,
			HmacEncoded:          config.AcmeHmac,
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

// FileExists checks whether a file exists.
func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// LoadPrivateKey loads a private key from a file.
func LoadPrivateKey(file string) (crypto.PrivateKey, error) {
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
