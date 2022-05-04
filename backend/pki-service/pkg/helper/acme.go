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

func RegisterAcme(client *lego.Client, config *cfg.SectigoConfiguration, account User, accountFile string, keyFile string) error {
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
